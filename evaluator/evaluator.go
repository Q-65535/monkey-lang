package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalStatements(node.Statements, env)
	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		if node.Value {
			return TRUE
		} else {
			return FALSE
		}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		return evalPrefix(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		right := Eval(node.Right, env)
		return evalInfix(node.Operator, left, right)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.ReturnStatement:
		return &object.ReturnValue{Value: Eval(node.ReturnValue, env)}
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.LetStatement:
		obj := Eval(node.Value, env)
		if obj.Type() == object.ERROR_OBJ {
			return obj
		}
		identStr := node.Name.Value
		env.Set(identStr, obj)
		return obj
	case *ast.FunctionLiteral:
		return &object.Function{Parameters: node.Parameters, Body: node.Body, Env: env}
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	default:
		return NULL
	}
}

func evalIdentifier(ident *ast.Identifier, env *object.Environment) object.Object {
	obj, ok := env.Get(ident.Value)
	if ok {
		return obj
	}
	obj, ok = builtinFuns[ident.Value]
	if ok {
		return obj
	}
	return newError("identifier '%s' not bind to any expression", ident.Value)
}

func evalCallExpression(call *ast.CallExpression, env *object.Environment) object.Object {
	function := Eval(call.Function, env)
	args := evalArgs(call.Arguments, env)

	switch fun := function.(type) {
	case *object.Error:
		return fun
	case *object.Builtin:
		return fun.Fn(args...)
	case *object.Function:
		if len(args) == 1 && args[0].Type() == object.ERROR_OBJ {
			return args[0]
		}
		functionEnv := object.NewCloseEnvironment(env)
		for i, arg := range args {
			curParam := fun.Parameters[i]
			functionEnv.Set(curParam.String(), arg)
		}
		resObj := Eval(fun.Body, functionEnv)
		if resObj.Type() == object.RETURN_VALUE_OBJ {
			r, _ := resObj.(*object.ReturnValue)
			resObj = r.Value
		}
		return resObj
	default:
		return newError("not a function reference: %s", function.Type())

	}
}

func evalArgs(exps []ast.Expression, env *object.Environment) []object.Object {
	args := []object.Object{}
	for _, exp := range exps {
		obj := Eval(exp, env)
		if obj.Type() == object.ERROR_OBJ {
			return []object.Object{obj}
		}
		args = append(args, Eval(exp, env))
	}
	return args
}

// @TODO: block statement should also be evaluated in a closure
func evalStatements(statements []ast.Statement, env *object.Environment) object.Object {
	var obj object.Object
	for _, st := range statements {
		obj = Eval(st, env)
		if obj.Type() == object.RETURN_VALUE_OBJ || obj.Type() == object.ERROR_OBJ {
			return obj
		}
	}
	return obj
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{ErrorMessage: fmt.Sprintf(format, a...)}
}

func evalPrefix(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return reverseBooleanize(right)
	case "-":
		return negate(right)
	default:
		return newError("invalid prefix: %s", operator)
	}
}

func nativeBool2Object(v bool) *object.Boolean {
	if v {
		return TRUE
	} else {
		return FALSE
	}
}

func reverseBooleanize(obj object.Object) object.Object {
	if obj == FALSE || obj == NULL {
		return TRUE
	} else {
		return FALSE
	}
}

func negate(obj object.Object) object.Object {
	if obj.Type() != object.INTEGER_OBJ {
		return NULL
	}
	val := obj.(*object.Integer).Value
	return &object.Integer{Value: -val}
}

func evalInfix(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfix(operator, left, right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfix(operator, left, right)
	case left.Type() != right.Type():
		return newError("mismatching type in infix:  %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("eval infix error:  %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBooleanInfix(operator string, left object.Object, right object.Object) object.Object {
	switch operator {
	case "==":
		return nativeBool2Object(left == right)
	case "!=":
		return nativeBool2Object(left != right)
	default:
		return newError("eval boolean infix error:  %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfix(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "==":
		return nativeBool2Object(leftVal == rightVal)
	case "!=":
		return nativeBool2Object(leftVal != rightVal)
	case "<":
		return nativeBool2Object(leftVal < rightVal)
	case ">":
		return nativeBool2Object(leftVal > rightVal)
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	default:
		return newError("eval integer infix error:  %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Altenative != nil {
		return Eval(ie.Altenative, env)
	} else {
		return newError("eval condition error, bad condition:  %s",
			condition.Type())
	}
}

func isTruthy(obj object.Object) bool {
	if obj == NULL || obj == FALSE {
		return false
	} else {
		return true
	}
}
