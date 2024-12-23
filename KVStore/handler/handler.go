package handler

import (
	"net/http"

	"github.com/SubhamMurarka/KVStore/models"
	"github.com/SubhamMurarka/KVStore/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type handler struct {
	repo repository.RepoInterface
}

func NewHandler(r repository.RepoInterface) *handler {
	return &handler{
		repo: r,
	}
}

func (h *handler) Put(c *gin.Context) {
	var ip *models.Request

	if err := c.ShouldBindJSON(&ip); err != nil {
		logrus.Error("Error Binding : ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err := h.repo.Put(ip)
	if err != nil {
		logrus.Error("Not able to PUT : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	logrus.Infof("Success for key %v", ip.Key)

	c.JSON(http.StatusOK, gin.H{"Success": "OK"})
}

func (h *handler) Get(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		logrus.Error("Key is missing in the query parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	ip, err := h.repo.Get(key)
	if err != nil {
		logrus.Error("Not able to GET : ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if ip == nil {
		logrus.Warn("No data found for key: ", key)
		c.JSON(http.StatusNotFound, gin.H{"error": "No data found"})
		return
	}

	logrus.Infof("Success for key %v", ip.Key)
	c.JSON(http.StatusOK, gin.H{"data": ip})
}

func (h *handler) Delete(c *gin.Context) {
	key := c.Query("key") // Use query parameter for key
	if key == "" {
		logrus.Error("Key is missing in the query parameter")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key is required"})
		return
	}

	if err := h.repo.Delete(key); err != nil {
		logrus.Error("Not able to DELETE: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logrus.Infof("Success for key %v", key)
	c.JSON(http.StatusOK, gin.H{"Success": "OK"})
}
