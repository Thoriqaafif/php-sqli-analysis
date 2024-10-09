package SSAGenerator

import "github.com/z7zmey/php-parser/pkg/ast"

type SSAGenerator struct {}

func (v *SSAGenerator) Root(n *ast.Root){}

func (v *SSAGenerator) Nullable(n *ast.Nullable){}

func (v *SSAGenerator) Parameter(n *ast.Parameter){}

func (v *SSAGenerator) Identifier(n *ast.Identifier){}

func (v *SSAGenerator) Argument(n *ast.Argument){}

func (v *SSAGenerator) Attribute(n *ast.Attribute){}

func (v *SSAGenerator) AttributeGroup(n *ast.AttributeGroup){}

func (v *SSAGenerator) StmtBreak(n *ast.StmtBreak){}

func (v *SSAGenerator) StmtCase(n *ast.StmtCase){}

func (v *SSAGenerator) StmtCatch(n *ast.StmtCatch){}

func (v *SSAGenerator) StmtClass(n *ast.StmtClass){}

func (v *SSAGenerator) StmtClassConstList(n *ast.StmtClassConstList){}

func (v *SSAGenerator) StmtClassMethod(n *ast.StmtClassMethod){}

func (v *SSAGenerator) StmtConstList(n *ast.StmtConstList){}

func (v *SSAGenerator) StmtConstant(n *ast.StmtConstant){}

func (v *SSAGenerator) StmtContinue(n *ast.StmtContinue){}

func (v *SSAGenerator) StmtDeclare(n *ast.StmtDeclare){}

func (v *SSAGenerator) StmtDefault(n *ast.StmtDefault){}

func (v *SSAGenerator) StmtDo(n *ast.StmtDo){}

func (v *SSAGenerator) StmtEcho(n *ast.StmtEcho){}

func (v *SSAGenerator) StmtElse(n *ast.StmtElse){}

func (v *SSAGenerator) StmtElseIf(n *ast.StmtElseIf){}

func (v *SSAGenerator) StmtExpression(n *ast.StmtExpression){}

func (v *SSAGenerator) StmtFinally(n *ast.StmtFinally){}

func (v *SSAGenerator) StmtFor(n *ast.StmtFor){}

func (v *SSAGenerator) StmtForeach(n *ast.StmtForeach){}

func (v *SSAGenerator) StmtFunction(n *ast.StmtFunction){}

func (v *SSAGenerator) StmtGlobal(n *ast.StmtGlobal){}

func (v *SSAGenerator) StmtGoto(n *ast.StmtGoto){}

func (v *SSAGenerator) StmtHaltCompiler(n *ast.StmtHaltCompiler){}

func (v *SSAGenerator) StmtIf(n *ast.StmtIf){}

func (v *SSAGenerator) StmtInlineHtml(n *ast.StmtInlineHtml){}

func (v *SSAGenerator) StmtInterface(n *ast.StmtInterface){}

func (v *SSAGenerator) StmtLabel(n *ast.StmtLabel){}

func (v *SSAGenerator) StmtNamespace(n *ast.StmtNamespace){}

func (v *SSAGenerator) StmtNop(n *ast.StmtNop){}

func (v *SSAGenerator) StmtProperty(n *ast.StmtProperty){}

func (v *SSAGenerator) StmtPropertyList(n *ast.StmtPropertyList){}

func (v *SSAGenerator) StmtReturn(n *ast.StmtReturn){}

func (v *SSAGenerator) StmtStatic(n *ast.StmtStatic){}

func (v *SSAGenerator) StmtStaticVar(n *ast.StmtStaticVar){}

func (v *SSAGenerator) StmtStmtList(n *ast.StmtStmtList){}

func (v *SSAGenerator) StmtSwitch(n *ast.StmtSwitch){}

func (v *SSAGenerator) StmtThrow(n *ast.StmtThrow){}

func (v *SSAGenerator) StmtTrait(n *ast.StmtTrait){}

func (v *SSAGenerator) StmtTraitUse(n *ast.StmtTraitUse){}

func (v *SSAGenerator) StmtTraitUseAlias(n *ast.StmtTraitUseAlias){}

func (v *SSAGenerator) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence){}

func (v *SSAGenerator) StmtTry(n *ast.StmtTry){}

func (v *SSAGenerator) StmtUnset(n *ast.StmtUnset){}

func (v *SSAGenerator) StmtUse(n *ast.StmtUseList){}

func (v *SSAGenerator) StmtGroupUse(n *ast.StmtGroupUseList){}

func (v *SSAGenerator) StmtUseDeclaration(n *ast.StmtUse){}

func (v *SSAGenerator) StmtWhile(n *ast.StmtWhile){}

func (v *SSAGenerator) ExprArray(n *ast.ExprArray){}

func (v *SSAGenerator) ExprArrayDimFetch(n *ast.ExprArrayDimFetch){}

func (v *SSAGenerator) ExprArrayItem(n *ast.ExprArrayItem){}

func (v *SSAGenerator) ExprArrowFunction(n *ast.ExprArrowFunction){}

func (v *SSAGenerator) ExprBrackets(n *ast.ExprBrackets){}

func (v *SSAGenerator) ExprBitwiseNot(n *ast.ExprBitwiseNot){}

func (v *SSAGenerator) ExprBooleanNot(n *ast.ExprBooleanNot){}

func (v *SSAGenerator) ExprClassConstFetch(n *ast.ExprClassConstFetch){}

func (v *SSAGenerator) ExprClone(n *ast.ExprClone){}

func (v *SSAGenerator) ExprClosure(n *ast.ExprClosure){}

func (v *SSAGenerator) ExprClosureUse(n *ast.ExprClosureUse){}

func (v *SSAGenerator) ExprConstFetch(n *ast.ExprConstFetch){}

func (v *SSAGenerator) ExprEmpty(n *ast.ExprEmpty){}

func (v *SSAGenerator) ExprErrorSuppress(n *ast.ExprErrorSuppress){}

func (v *SSAGenerator) ExprEval(n *ast.ExprEval){}

func (v *SSAGenerator) ExprExit(n *ast.ExprExit){}

func (v *SSAGenerator) ExprFunctionCall(n *ast.ExprFunctionCall){}

func (v *SSAGenerator) ExprInclude(n *ast.ExprInclude){}

func (v *SSAGenerator) ExprIncludeOnce(n *ast.ExprIncludeOnce){}

func (v *SSAGenerator) ExprInstanceOf(n *ast.ExprInstanceOf){}

func (v *SSAGenerator) ExprIsset(n *ast.ExprIsset){}

func (v *SSAGenerator) ExprList(n *ast.ExprList){}

func (v *SSAGenerator) ExprMethodCall(n *ast.ExprMethodCall){}

func (v *SSAGenerator) ExprNew(n *ast.ExprNew){}

func (v *SSAGenerator) ExprPostDec(n *ast.ExprPostDec){}

func (v *SSAGenerator) ExprPostInc(n *ast.ExprPostInc){}

func (v *SSAGenerator) ExprPreDec(n *ast.ExprPreDec){}

func (v *SSAGenerator) ExprPreInc(n *ast.ExprPreInc){}

func (v *SSAGenerator) ExprPrint(n *ast.ExprPrint){}

func (v *SSAGenerator) ExprPropertyFetch(n *ast.ExprPropertyFetch){}

func (v *SSAGenerator) ExprRequire(n *ast.ExprRequire){}

func (v *SSAGenerator) ExprRequireOnce(n *ast.ExprRequireOnce){}

func (v *SSAGenerator) ExprShellExec(n *ast.ExprShellExec){}

func (v *SSAGenerator) ExprStaticCall(n *ast.ExprStaticCall){}

func (v *SSAGenerator) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch){}

func (v *SSAGenerator) ExprTernary(n *ast.ExprTernary){}

func (v *SSAGenerator) ExprUnaryMinus(n *ast.ExprUnaryMinus){}

func (v *SSAGenerator) ExprUnaryPlus(n *ast.ExprUnaryPlus){}

func (v *SSAGenerator) ExprVariable(n *ast.ExprVariable){}

func (v *SSAGenerator) ExprYield(n *ast.ExprYield){}

func (v *SSAGenerator) ExprYieldFrom(n *ast.ExprYieldFrom){}

func (v *SSAGenerator) ExprAssign(n *ast.ExprAssign){}

func (v *SSAGenerator) ExprAssignReference(n *ast.ExprAssignReference){}

func (v *SSAGenerator) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd){}

func (v *SSAGenerator) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr){}

func (v *SSAGenerator) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor){}

func (v *SSAGenerator) ExprAssignCoalesce(n *ast.ExprAssignCoalesce){}

func (v *SSAGenerator) ExprAssignConcat(n *ast.ExprAssignConcat){}

func (v *SSAGenerator) ExprAssignDiv(n *ast.ExprAssignDiv){}

func (v *SSAGenerator) ExprAssignMinus(n *ast.ExprAssignMinus){}

func (v *SSAGenerator) ExprAssignMod(n *ast.ExprAssignMod){}

func (v *SSAGenerator) ExprAssignMul(n *ast.ExprAssignMul){}

func (v *SSAGenerator) ExprAssignPlus(n *ast.ExprAssignPlus){}

func (v *SSAGenerator) ExprAssignPow(n *ast.ExprAssignPow){}

func (v *SSAGenerator) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft){}

func (v *SSAGenerator) ExprAssignShiftRight(n *ast.ExprAssignShiftRight){}

func (v *SSAGenerator) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd){}

func (v *SSAGenerator) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr){}

func (v *SSAGenerator) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor){}

func (v *SSAGenerator) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd){}

func (v *SSAGenerator) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr){}

func (v *SSAGenerator) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce){}

func (v *SSAGenerator) ExprBinaryConcat(n *ast.ExprBinaryConcat){}

func (v *SSAGenerator) ExprBinaryDiv(n *ast.ExprBinaryDiv){}

func (v *SSAGenerator) ExprBinaryEqual(n *ast.ExprBinaryEqual){}

func (v *SSAGenerator) ExprBinaryGreater(n *ast.ExprBinaryGreater){}

func (v *SSAGenerator) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual){}

func (v *SSAGenerator) ExprBinaryIdentical(n *ast.ExprBinaryIdentical){}

func (v *SSAGenerator) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd){}

func (v *SSAGenerator) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr){}

func (v *SSAGenerator) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor){}

func (v *SSAGenerator) ExprBinaryMinus(n *ast.ExprBinaryMinus){}

func (v *SSAGenerator) ExprBinaryMod(n *ast.ExprBinaryMod){}

func (v *SSAGenerator) ExprBinaryMul(n *ast.ExprBinaryMul){}

func (v *SSAGenerator) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual){}

func (v *SSAGenerator) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical){}

func (v *SSAGenerator) ExprBinaryPlus(n *ast.ExprBinaryPlus){}

func (v *SSAGenerator) ExprBinaryPow(n *ast.ExprBinaryPow){}

func (v *SSAGenerator) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft){}

func (v *SSAGenerator) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight){}

func (v *SSAGenerator) ExprBinarySmaller(n *ast.ExprBinarySmaller){}

func (v *SSAGenerator) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual){}

func (v *SSAGenerator) ExprBinarySpaceship(n *ast.ExprBinarySpaceship){}

func (v *SSAGenerator) ExprCastArray(n *ast.ExprCastArray){}

func (v *SSAGenerator) ExprCastBool(n *ast.ExprCastBool){}

func (v *SSAGenerator) ExprCastDouble(n *ast.ExprCastDouble){}

func (v *SSAGenerator) ExprCastInt(n *ast.ExprCastInt){}

func (v *SSAGenerator) ExprCastObject(n *ast.ExprCastObject){}

func (v *SSAGenerator) ExprCastString(n *ast.ExprCastString){}

func (v *SSAGenerator) ExprCastUnset(n *ast.ExprCastUnset){}

func (v *SSAGenerator) ScalarDnumber(n *ast.ScalarDnumber){}

func (v *SSAGenerator) ScalarEncapsed(n *ast.ScalarEncapsed){}

func (v *SSAGenerator) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart){}

func (v *SSAGenerator) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar){}

func (v *SSAGenerator) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets){}

func (v *SSAGenerator) ScalarHeredoc(n *ast.ScalarHeredoc){}

func (v *SSAGenerator) ScalarLnumber(n *ast.ScalarLnumber){}

func (v *SSAGenerator) ScalarMagicConstant(n *ast.ScalarMagicConstant){}

func (v *SSAGenerator) ScalarString(n *ast.ScalarString){}

func (v *SSAGenerator) NameName(n *ast.Name){}

func (v *SSAGenerator) NameFullyQualified(n *ast.NameFullyQualified){}

func (v *SSAGenerator) NameRelative(n *ast.NameRelative){}

func (v *SSAGenerator) NameNamePart(n *ast.NamePart){}

