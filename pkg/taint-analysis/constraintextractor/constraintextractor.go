package constraintextractor

import (
	"github.com/Thoriqaafif/php-sqli-analysis/pkg/taint-analysis/cfg"

	"github.com/aclements/go-z3/z3"
)

func ExtractConstraints(ctx *z3.Context, oper cfg.Operand) z3.Bool {
	return ctx.FromBool(false)
}
