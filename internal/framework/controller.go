package thinkgo

// Controller is the base controller that other controllers embed.
// Like ThinkPHP's BaseController.
//
// Usage:
//
//	type UserController struct {
//	    thinkgo.Controller
//	}
//
//	func (c *UserController) Index(ctx *thinkgo.Context) error {
//	    return ctx.JSON(thinkgo.Map{"code": 1, "data": "hello"})
//	}
type Controller struct {
	App *App
}

// Prepare is called before every action.
// Override this method to implement initialization logic
// (like ThinkPHP's _initialize()).
func (c *Controller) Prepare(ctx *Context) error {
	return nil
}

// Finish is called after every action.
// Override this method to implement cleanup logic.
func (c *Controller) Finish(ctx *Context) error {
	return nil
}

// JSON sends a JSON response.
func (c *Controller) JSON(ctx *Context, data any) error {
	return NewResponse(ctx).JSON(data)
}

// Success sends a ThinkPHP-style success response.
func (c *Controller) Success(ctx *Context, msg string, data ...any) error {
	return NewResponse(ctx).Success(msg, data...)
}

// Error sends a ThinkPHP-style error response.
func (c *Controller) Error(ctx *Context, msg string, data ...any) error {
	return NewResponse(ctx).Error(msg, data...)
}

// Map is a shorthand for map[string]any.
type Map = map[string]any
