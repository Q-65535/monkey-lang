package evaluator

import "monkey/object"

var builtinFuns = map[string]*object.Builtin{
	"len": &object.Builtin{Fn: builtinLen},
}

func builtinLen(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("len(): expect 1 arguments, but got %d", len(args))
	}
	switch obj := args[0].(type) {
	case *object.Integer:
		return &object.Integer{Value: 32}
	case *object.String:
		return &object.Integer{Value: int64(len(obj.Value))}
	case *object.Array:
		length := len(obj.Value)
		return &object.Integer{Value: int64(length)}
	default:
		return newError("len (currently) doesn't support %s type", obj.Type())
	}
}
