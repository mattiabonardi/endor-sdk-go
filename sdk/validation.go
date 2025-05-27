package sdk

func ValidationHandler[T any](c *EndorContext[T]) {
	if err := c.GinContext.ShouldBindJSON(&c.Payload); err != nil {
		c.BadRequest(err)
		return
	}
	c.Next()
}
