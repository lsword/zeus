package httpserver

import (
	//"encoding/json"
	"fmt"
	//"io"
	//"io/ioutil"
	//"net"
	//"net/http"
	_ "net/http/pprof"

	//"github.com/golang/net/netutil"
	//"github.com/gorilla/mux"
	//"github.com/kataras/iris"
	"github.com/gin-gonic/gin"
	//"github.com/thoas/stats"

	"zeus/config"
	//"goserver/log"
	//"goserver/pkg/version"
)

//var middleware *stats.Stats

func InitRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.GET("/serverinfo", getServerInfo)
	//router.GET("/serverstats", getServerStats)
	//router.GET("/serverconfig", getServerConfig)
	//router.GET("/testquery", getTestQuery)
	//router.GET("/testexec", getTestExec)
	return router
}

/*
func getServerInfo(c *gin.Context) {
	c.String(http.StatusOK, dbserver.Status())
}

func getServerStats(c *gin.Context) {
	c.String(http.StatusOK, dbserver.Status())
}

func getServerConfig(c *gin.Context) {
	c.String(http.StatusOK, fmt.Sprintf("%s", config.ConfigJson()))
}

func getTestQuery(c *gin.Context) {
	db := dbserver.GetDatabase()
	_, rows, _ := db.QueryData("mysql1", "select * from test where id=5")
	result := ""
	for row := range *rows {
		for k, v := range (*rows)[row] {
			if v == nil {
				result = result + k + ":NULL" + "; "
				continue
			}
			switch k {
			case "id":
				result = result + fmt.Sprintf("%s:%s;", k, string(v.([]byte)))
				break
			case "name":
				result = result + fmt.Sprintf("%s:%s;", k, string(v.([]byte)))
				break
			}
		}
		result = result + "\n"
	}
	c.String(http.StatusOK, result)
}

func getTestExec(c *gin.Context) {
	db := dbserver.GetDatabase()
	lastid, affectrow, err := db.Exec("mysql1", "insert into test(id,name) values (?,?)", 3, "123")
	if err != nil {
		c.String(http.StatusOK, fmt.Sprintf("lastid:%d, affectrow:%d, error:%s", lastid, affectrow, err.Error()))
	} else {
		c.String(http.StatusOK, fmt.Sprintf("lastid:%d, affectrow:%d", lastid, affectrow))
	}

}
*/

func Run(c *config.Config) {
	if c.HttpServer.Switch != "on" {
		return
	}

	//iris.UseFunc()
	//middleware = stats.New()
	//iris.Use(stats)
	router := InitRouter()

	go router.Run(fmt.Sprintf(":%d", c.HttpServer.Port))
}
