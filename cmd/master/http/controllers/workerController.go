package controllers

import (
	"github.com/Jake4-CX/CT6039-Dissertation-Backend-Test-2/cmd/master/managers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetWorkers(c *gin.Context) {
	workers := managers.GetAvailableWorkers()
	c.JSON(http.StatusOK, workers)
}