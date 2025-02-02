package handlers

import (
    //"net/http"

    "github.com/gin-gonic/gin"
)


func MainPage(c *gin.Context) {
    response := map[int]int{5: 5}
        c.JSON(200, response)
} 

func Register(c *gin.Context){

}