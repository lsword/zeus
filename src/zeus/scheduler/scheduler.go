package scheduler

import (
	"flag"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/auth"
	"github.com/mesos/mesos-go/auth/sasl"
	//"github.com/mesos/mesos-go/auth/sasl/mech"
	mesos "github.com/mesos/mesos-go/mesosproto"
	util "github.com/mesos/mesos-go/mesosutil"
	sched "github.com/mesos/mesos-go/scheduler"
	"golang.org/x/net/context"

	"zeus/config"
	"zeus/log"
)

const (
	CPUS_PER_TASK = 1
	MEM_PER_TASK  = 128
)

type Scheduler struct {
	master        string
	framework     *mesos.FrameworkInfo
	tasksLaunched int
	tasksFinished int
	tasksErrored  int
}

func NewScheduler(master string, fwInfo *mesos.FrameworkInfo) *Scheduler {
	return &Scheduler{
		master:        master,
		framework:     fwInfo,
		tasksLaunched: 0,
		tasksFinished: 0,
		tasksErrored:  0,
	}
}

func Run(c *config.Config) {
	// the framework
	fwinfo := &mesos.FrameworkInfo{
		User: proto.String(""), // Mesos-go will fill in user.
		Name: proto.String("Zeus Framework (Go)"),
	}

	cred := (*mesos.Credential)(nil)
	if c.Scheduler.AuthPrincipal != "" {
		fwinfo.Principal = proto.String(c.Scheduler.AuthPrincipal)
		cred = &mesos.Credential{
			Principal: proto.String(c.Scheduler.AuthPrincipal),
		}
		if c.Scheduler.AuthSecretFile != "" {
			_, err := os.Stat(c.Scheduler.AuthSecretFile)
			if err != nil {
				log.Fatal("missing secret file: ", err.Error())
			}
			secret, err := ioutil.ReadFile(c.Scheduler.AuthSecretFile)
			if err != nil {
				log.Fatal("failed to read secret file: ", err.Error())
			}
			cred.Secret = proto.String(string(secret))
		}
	}

	bindingAddress := parseIP(c.Scheduler.Address)
	config := sched.DriverConfig{
		Scheduler:  NewScheduler(c.Scheduler.Master, fwinfo),
		Framework:  fwinfo,
		Master:     c.Scheduler.Master,
		Credential: cred,
		WithAuthContext: func(ctx context.Context) context.Context {
			ctx = auth.WithLoginProvider(ctx, sasl.ProviderName)
			ctx = sasl.WithBindingAddress(ctx, bindingAddress)
			return ctx
		},
	}
	driver, err := sched.NewMesosSchedulerDriver(config)

	if err != nil {
		log.Errorf("Unable to create a SchedulerDriver ", err.Error())
	}

	if stat, err := driver.Run(); err != nil {
		log.Infof("Framework stopped with status %s and error: %s\n", stat.String(), err.Error())
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}
	log.Infof("framework terminating")
}

func (sched *Scheduler) Registered(driver sched.SchedulerDriver, frameworkId *mesos.FrameworkID, masterInfo *mesos.MasterInfo) {
	log.Infof("Framework Registered with Master ", masterInfo)
}

func (sched *Scheduler) Reregistered(driver sched.SchedulerDriver, masterInfo *mesos.MasterInfo) {
	log.Infof("Framework Re-Registered with Master ", masterInfo)
	_, err := driver.ReconcileTasks([]*mesos.TaskStatus{})
	if err != nil {
		log.Errorf("failed to request task reconciliation: %v", err)
	}
}

func (sched *Scheduler) Disconnected(sched.SchedulerDriver) {
	log.Warnf("disconnected from master")
}

func (sched *Scheduler) ResourceOffers(driver sched.SchedulerDriver, offers []*mesos.Offer) {

	if (sched.tasksLaunched - sched.tasksErrored) >= 5 {
		log.Info("decline all of the offers since all of our tasks are already launched")
		ids := make([]*mesos.OfferID, len(offers))
		for i, offer := range offers {
			ids[i] = offer.Id
		}
		driver.LaunchTasks(ids, []*mesos.TaskInfo{}, &mesos.Filters{RefuseSeconds: proto.Float64(120)})
		return
	}
	for _, offer := range offers {
		cpuResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "cpus"
		})
		cpus := 0.0
		for _, res := range cpuResources {
			cpus += res.GetScalar().GetValue()
		}

		memResources := util.FilterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "mem"
		})
		mems := 0.0
		for _, res := range memResources {
			mems += res.GetScalar().GetValue()
		}

		log.Infof("Received Offer <", offer.Id.GetValue(), "> with cpus=", cpus, " mem=", mems)

		remainingCpus := cpus
		remainingMems := mems

		// account for executor resources if there's not an executor already running on the slave
		if len(offer.ExecutorIds) == 0 {
			remainingCpus -= 1
			remainingMems -= 128
		}

		var tasks []*mesos.TaskInfo
		for (sched.tasksLaunched-sched.tasksErrored) < 5 &&
			CPUS_PER_TASK <= remainingCpus &&
			MEM_PER_TASK <= remainingMems {

			sched.tasksLaunched++

			taskId := &mesos.TaskID{
				Value: proto.String(strconv.Itoa(sched.tasksLaunched)),
			}

			task := &mesos.TaskInfo{
				Name:    proto.String("go-task-" + taskId.GetValue()),
				TaskId:  taskId,
				SlaveId: offer.SlaveId,
				//Executor: sched.executor,
				Resources: []*mesos.Resource{
					util.NewScalarResource("cpus", CPUS_PER_TASK),
					util.NewScalarResource("mem", MEM_PER_TASK),
				},
			}
			log.Infof("Prepared task: %s with offer %s for launch\n", task.GetName(), offer.Id.GetValue())

			tasks = append(tasks, task)
			remainingCpus -= CPUS_PER_TASK
			remainingMems -= MEM_PER_TASK
		}
		log.Infof("Launching ", len(tasks), "tasks for offer", offer.Id.GetValue())
		driver.LaunchTasks([]*mesos.OfferID{offer.Id}, tasks, &mesos.Filters{RefuseSeconds: proto.Float64(5)})
	}
}

func (sched *Scheduler) StatusUpdate(driver sched.SchedulerDriver, status *mesos.TaskStatus) {
	log.Infof("Status update: task", status.TaskId.GetValue(), " is in state ", status.State.Enum().String())
	if status.GetState() == mesos.TaskState_TASK_FINISHED {
		sched.tasksFinished++
		driver.ReviveOffers() // TODO(jdef) rate-limit this
	}

	if sched.tasksFinished >= 5 {
		log.Infof("Total tasks completed, stopping framework.")
		driver.Stop(false)
	}

	if status.GetState() == mesos.TaskState_TASK_LOST ||
		status.GetState() == mesos.TaskState_TASK_KILLED ||
		status.GetState() == mesos.TaskState_TASK_FAILED ||
		status.GetState() == mesos.TaskState_TASK_ERROR {
		sched.tasksErrored++
	}
}

func (sched *Scheduler) OfferRescinded(_ sched.SchedulerDriver, oid *mesos.OfferID) {
	log.Errorf("offer rescinded: %v", oid)
}
func (sched *Scheduler) FrameworkMessage(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, msg string) {
	log.Errorf("framework message from executor %q slave %q: %q", eid, sid, msg)
}
func (sched *Scheduler) SlaveLost(_ sched.SchedulerDriver, sid *mesos.SlaveID) {
	log.Errorf("slave lost: %v", sid)
}
func (sched *Scheduler) ExecutorLost(_ sched.SchedulerDriver, eid *mesos.ExecutorID, sid *mesos.SlaveID, code int) {
	log.Errorf("executor %q lost on slave %q code %d", eid, sid, code)
}
func (sched *Scheduler) Error(_ sched.SchedulerDriver, err string) {
	log.Errorf("Scheduler received error: %v", err)
}

// ----------------------- func init() ------------------------- //

func init() {
	flag.Parse()
	log.Infof("Initializing the Example Scheduler...")
}

// returns (downloadURI, basename(path))
/*
func serveExecutorArtifact(path string) (*string, string) {
	serveFile := func(pattern string, filename string) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filename)
		})
	}

	// Create base path (http://foobar:5000/<base>)
	pathSplit := strings.Split(path, "/")
	var base string
	if len(pathSplit) > 0 {
		base = pathSplit[len(pathSplit)-1]
	} else {
		base = path
	}
	serveFile("/"+base, path)

	hostURI := fmt.Sprintf("http://%s:%d/%s", *address, *artifactPort, base)
	log.Infof("Hosting artifact '%s' at '%s'", path, hostURI)

	return &hostURI, base
}

func prepareExecutorInfo() *mesos.ExecutorInfo {
	executorUris := []*mesos.CommandInfo_URI{}
	uri, executorCmd := serveExecutorArtifact(*executorPath)
	executorUris = append(executorUris, &mesos.CommandInfo_URI{Value: uri, Executable: proto.Bool(true)})

	// forward the value of the scheduler's -v flag to the executor
	v := 0
	if f := flag.Lookup("v"); f != nil && f.Value != nil {
		if vstr := f.Value.String(); vstr != "" {
			if vi, err := strconv.ParseInt(vstr, 10, 32); err == nil {
				v = int(vi)
			}
		}
	}
	executorCommand := fmt.Sprintf("./%s -logtostderr=true -v=%d -slow_tasks=%v", executorCmd, v, *slowTasks)

	go http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *artifactPort), nil)
	log.V(2).Info("Serving executor artifacts...")

	// Create mesos scheduler driver.
	return &mesos.ExecutorInfo{
		ExecutorId: util.NewExecutorID("default"),
		Name:       proto.String("Test Executor (Go)"),
		Source:     proto.String("go_test"),
		Command: &mesos.CommandInfo{
			Value: proto.String(executorCommand),
			Uris:  executorUris,
		},
		Resources: []*mesos.Resource{
			util.NewScalarResource("cpus", CPUS_PER_EXECUTOR),
			util.NewScalarResource("mem", MEM_PER_EXECUTOR),
		},
	}
}

*/
func parseIP(address string) net.IP {
	addr, err := net.LookupIP(address)
	if err != nil {
		log.Fatal(err)
	}
	if len(addr) < 1 {
		log.Fatalf("failed to parse IP from address '%v'", address)
	}
	return addr[0]
}
