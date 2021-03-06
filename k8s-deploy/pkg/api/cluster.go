package api

import (
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/db"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/job"
	"github.com/FederatedAI/KubeFATE/k8s-deploy/pkg/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Cluster struct {
}

// Router is cluster router definition method
func (c *Cluster) Router(r *gin.RouterGroup) {

	authMiddleware, _ := GetAuthMiddleware()
	cluster := r.Group("/cluster")
	cluster.Use(authMiddleware.MiddlewareFunc())
	{
		cluster.POST("", c.createCluster)
		cluster.PUT("", c.setCluster)
		cluster.GET("/", c.getClusterList)
		cluster.GET("/:clusterId", c.getCluster)
		cluster.DELETE("/:clusterId", c.deleteCluster)
	}
}

func (_ *Cluster) createCluster(c *gin.Context) {

	user, _ := c.Get(identityKey)

	clusterArgs := new(job.ClusterArgs)

	if err := c.ShouldBindJSON(&clusterArgs); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Debug().Interface("parameters", clusterArgs).Msg("parameters")

	// create job and use goroutine do job result save to db
	j, err := job.ClusterInstall(clusterArgs, user.(*User).Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "createCluster success", "data": j})
}

func (_ *Cluster) setCluster(c *gin.Context) {

	//cluster := new(db.Cluster)
	//if err := c.ShouldBindJSON(&cluster); err != nil {
	//	c.JSON(400, gin.H{"error": err.Error()})
	//	return
	//}

	user, _ := c.Get(identityKey)

	clusterArgs := new(job.ClusterArgs)

	if err := c.ShouldBindJSON(&clusterArgs); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	log.Debug().Interface("parameters", clusterArgs).Msg("parameters")

	// create job and use goroutine do job result save to db
	j, err := job.ClusterUpdate(clusterArgs, user.(*User).Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "setCluster success", "data": j})
}

func (_ *Cluster) getCluster(c *gin.Context) {

	clusterId := c.Param("clusterId")
	if clusterId == "" {
		c.JSON(400, gin.H{"error": "not exit clusterId"})
		return
	}

	//cluster, err := db.ClusterFindByName(clusterId, clusterId)
	//if err != nil {
	//	c.JSON(500, gin.H{"error": err})
	//	return
	//}

	cluster, err := db.ClusterFindByUUID(clusterId)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	ip, err := service.GetNodeIp()
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}
	port, err := service.GetProxySvcNodePorts(cluster.Name, cluster.NameSpace)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}
	podList, err := service.GetPodList(cluster.Name, cluster.NameSpace)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if cluster.Info == nil {
		cluster.Info = make(map[string]interface{})
	}

	if len(ip) > 0 {
		cluster.Info["ip"] = ip[len(ip)-1]
	}
	if len(port) > 0 {
		cluster.Info["port"] = port[0]
	}

	cluster.Info["modules"] = podList

	if cluster.ChartValues != nil {

		cluster.Info["dashboard"] = cluster.ChartValues["host"].(map[string]interface{})["fateboard"]
	}

	if cluster.Config == nil {
		cluster.Config = make(map[string]interface{})
	}

	c.JSON(200, gin.H{"data": cluster})
}

func (_ *Cluster) getClusterList(c *gin.Context) {

	all := false
	qall := c.Query("all")
	if qall == "true" {
		all = true
	}

	log.Debug().Bool("all", all).Msg("get args")

	clusterList, err := db.FindClusterList("", all)

	var clusterListreturn = make([]*db.Cluster, 0)
	if !all {
		for _, v := range clusterList {
			if v.Status != db.Deleted_c {
				clusterListreturn = append(clusterListreturn, v)
			}
		}
	} else {
		clusterListreturn = clusterList
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "getClusterList success", "data": clusterListreturn})
}

func (_ *Cluster) deleteCluster(c *gin.Context) {

	user, _ := c.Get(identityKey)

	clusterId := c.Param("clusterId")
	if clusterId == "" {
		c.JSON(400, gin.H{"error": "not exit clusterId"})
	}

	j, err := job.ClusterDelete(clusterId, user.(*User).Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "deleteCluster success", "data": j})
}
