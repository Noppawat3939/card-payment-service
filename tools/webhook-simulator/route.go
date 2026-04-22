package main

import (
	"card-payment-service/internal/response"

	"github.com/gin-gonic/gin"
)

func RegisterWebhookRoutes(r *gin.Engine, sim *Simulator) {
	// get available events
	r.GET("/events", func(c *gin.Context) {
		response.OK(c, availableEvents())
	})

	// batch send events
	r.POST("/simulate/batch", func(c *gin.Context) {
		var req []struct {
			Event      string `json:"event"`
			GatewayRef string `json:"gateway_ref"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c)
			return
		}

		results := make([]*SimulateResult, 0, len(req))
		for _, item := range req {
			result, err := sim.Send(EventType(item.Event), item.GatewayRef)
			if err != nil {
				results = append(results, &SimulateResult{
					Event:      item.Event,
					GatewayRef: item.GatewayRef,
					Success:    false,
					Error:      err.Error(),
				})
				continue
			}
			results = append(results, result)
		}

		response.OK(c, map[string][]*SimulateResult{"results": results})
	})

	// health check
	r.GET("/health", func(c *gin.Context) {
		response.OK(c, map[string]string{
			"status":     "ok",
			"target_url": sim.targetURL,
		})
	})
}
