package PathFinder

import "github.com/z7zmey/php-parser/pkg/ast"

type PathFinder struct{}

func (v *PathFinder) Root(n *ast.Root) {}

func (v *PathFinder) Nullable(n *ast.Nullable) {}

func (v *PathFinder) Parameter(n *ast.Parameter) {}

func (v *PathFinder) Identifier(n *ast.Identifier) {}

func (v *PathFinder) Argument(n *ast.Argument) {}

func (v *PathFinder) Attribute(n *ast.Attribute) {}

func (v *PathFinder) AttributeGroup(n *ast.AttributeGroup) {}

func (v *PathFinder) StmtBreak(n *ast.StmtBreak) {}

func (v *PathFinder) StmtCase(n *ast.StmtCase) {}

func (v *PathFinder) StmtCatch(n *ast.StmtCatch) {}

func (v *PathFinder) StmtClass(n *ast.StmtClass) {}

func (v *PathFinder) StmtClassConstList(n *ast.StmtClassConstList) {}

func (v *PathFinder) StmtClassMethod(n *ast.StmtClassMethod) {}

func (v *PathFinder) StmtConstList(n *ast.StmtConstList) {}

func (v *PathFinder) StmtConstant(n *ast.StmtConstant) {}

func (v *PathFinder) StmtContinue(n *ast.StmtContinue) {}

func (v *PathFinder) StmtDeclare(n *ast.StmtDeclare) {}

func (v *PathFinder) StmtDefault(n *ast.StmtDefault) {}

func (v *PathFinder) StmtDo(n *ast.StmtDo) {}

func (v *PathFinder) StmtEcho(n *ast.StmtEcho) {}

func (v *PathFinder) StmtElse(n *ast.StmtElse) {}

func (v *PathFinder) StmtElseIf(n *ast.StmtElseIf) {}

func (v *PathFinder) StmtExpression(n *ast.StmtExpression) {}

func (v *PathFinder) StmtFinally(n *ast.StmtFinally) {}

func (v *PathFinder) StmtFor(n *ast.StmtFor) {}

func (v *PathFinder) StmtForeach(n *ast.StmtForeach) {}

func (v *PathFinder) StmtFunction(n *ast.StmtFunction) {}

func (v *PathFinder) StmtGlobal(n *ast.StmtGlobal) {}

func (v *PathFinder) StmtGoto(n *ast.StmtGoto) {}

func (v *PathFinder) StmtHaltCompiler(n *ast.StmtHaltCompiler) {}

func (v *PathFinder) StmtIf(n *ast.StmtIf) {}

func (v *PathFinder) StmtInlineHtml(n *ast.StmtInlineHtml) {}

func (v *PathFinder) StmtInterface(n *ast.StmtInterface) {}

func (v *PathFinder) StmtLabel(n *ast.StmtLabel) {}

func (v *PathFinder) StmtNamespace(n *ast.StmtNamespace) {}

func (v *PathFinder) StmtNop(n *ast.StmtNop) {}

func (v *PathFinder) StmtProperty(n *ast.StmtProperty) {}

func (v *PathFinder) StmtPropertyList(n *ast.StmtPropertyList) {}

func (v *PathFinder) StmtReturn(n *ast.StmtReturn) {}

func (v *PathFinder) StmtStatic(n *ast.StmtStatic) {}

func (v *PathFinder) StmtStaticVar(n *ast.StmtStaticVar) {}

func (v *PathFinder) StmtStmtList(n *ast.StmtStmtList) {}

func (v *PathFinder) StmtSwitch(n *ast.StmtSwitch) {}

func (v *PathFinder) StmtThrow(n *ast.StmtThrow) {}

func (v *PathFinder) StmtTrait(n *ast.StmtTrait) {}

func (v *PathFinder) StmtTraitUse(n *ast.StmtTraitUse) {}

func (v *PathFinder) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {}

func (v *PathFinder) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {}

func (v *PathFinder) StmtTry(n *ast.StmtTry) {}

func (v *PathFinder) StmtUnset(n *ast.StmtUnset) {}

func (v *PathFinder) StmtUse(n *ast.StmtUseList) {}

func (v *PathFinder) StmtGroupUse(n *ast.StmtGroupUseList) {}

func (v *PathFinder) StmtUseDeclaration(n *ast.StmtUse) {}

func (v *PathFinder) StmtWhile(n *ast.StmtWhile) {}

func (v *PathFinder) ExprArray(n *ast.ExprArray) {}

func (v *PathFinder) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {}

func (v *PathFinder) ExprArrayItem(n *ast.ExprArrayItem) {}

func (v *PathFinder) ExprArrowFunction(n *ast.ExprArrowFunction) {}

func (v *PathFinder) ExprBrackets(n *ast.ExprBrackets) {}

func (v *PathFinder) ExprBitwiseNot(n *ast.ExprBitwiseNot) {}

func (v *PathFinder) ExprBooleanNot(n *ast.ExprBooleanNot) {}

func (v *PathFinder) ExprClassConstFetch(n *ast.ExprClassConstFetch) {}

func (v *PathFinder) ExprClone(n *ast.ExprClone) {}

func (v *PathFinder) ExprClosure(n *ast.ExprClosure) {}

func (v *PathFinder) ExprClosureUse(n *ast.ExprClosureUse) {}

func (v *PathFinder) ExprConstFetch(n *ast.ExprConstFetch) {}

func (v *PathFinder) ExprEmpty(n *ast.ExprEmpty) {}

func (v *PathFinder) ExprErrorSuppress(n *ast.ExprErrorSuppress) {}

func (v *PathFinder) ExprEval(n *ast.ExprEval) {}

func (v *PathFinder) ExprExit(n *ast.ExprExit) {}

func (v *PathFinder) ExprFunctionCall(n *ast.ExprFunctionCall) {}

func (v *PathFinder) ExprInclude(n *ast.ExprInclude) {}

func (v *PathFinder) ExprIncludeOnce(n *ast.ExprIncludeOnce) {}

func (v *PathFinder) ExprInstanceOf(n *ast.ExprInstanceOf) {}

func (v *PathFinder) ExprIsset(n *ast.ExprIsset) {}

func (v *PathFinder) ExprList(n *ast.ExprList) {}

func (v *PathFinder) ExprMethodCall(n *ast.ExprMethodCall) {}

func (v *PathFinder) ExprNew(n *ast.ExprNew) {}

func (v *PathFinder) ExprPostDec(n *ast.ExprPostDec) {}

func (v *PathFinder) ExprPostInc(n *ast.ExprPostInc) {}

func (v *PathFinder) ExprPreDec(n *ast.ExprPreDec) {}

func (v *PathFinder) ExprPreInc(n *ast.ExprPreInc) {}

func (v *PathFinder) ExprPrint(n *ast.ExprPrint) {}

func (v *PathFinder) ExprPropertyFetch(n *ast.ExprPropertyFetch) {}

func (v *PathFinder) ExprRequire(n *ast.ExprRequire) {}

func (v *PathFinder) ExprRequireOnce(n *ast.ExprRequireOnce) {}

func (v *PathFinder) ExprShellExec(n *ast.ExprShellExec) {}

func (v *PathFinder) ExprStaticCall(n *ast.ExprStaticCall) {}

func (v *PathFinder) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {}

func (v *PathFinder) ExprTernary(n *ast.ExprTernary) {}

func (v *PathFinder) ExprUnaryMinus(n *ast.ExprUnaryMinus) {}

func (v *PathFinder) ExprUnaryPlus(n *ast.ExprUnaryPlus) {}

func (v *PathFinder) ExprVariable(n *ast.ExprVariable) {}

func (v *PathFinder) ExprYield(n *ast.ExprYield) {}

func (v *PathFinder) ExprYieldFrom(n *ast.ExprYieldFrom) {}

func (v *PathFinder) ExprAssign(n *ast.ExprAssign) {}

func (v *PathFinder) ExprAssignReference(n *ast.ExprAssignReference) {}

func (v *PathFinder) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {}

func (v *PathFinder) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {}

func (v *PathFinder) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {}

func (v *PathFinder) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {}

func (v *PathFinder) ExprAssignConcat(n *ast.ExprAssignConcat) {}

func (v *PathFinder) ExprAssignDiv(n *ast.ExprAssignDiv) {}

func (v *PathFinder) ExprAssignMinus(n *ast.ExprAssignMinus) {}

func (v *PathFinder) ExprAssignMod(n *ast.ExprAssignMod) {}

func (v *PathFinder) ExprAssignMul(n *ast.ExprAssignMul) {}

func (v *PathFinder) ExprAssignPlus(n *ast.ExprAssignPlus) {}

func (v *PathFinder) ExprAssignPow(n *ast.ExprAssignPow) {}

func (v *PathFinder) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {}

func (v *PathFinder) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {}

func (v *PathFinder) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {}

func (v *PathFinder) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {}

func (v *PathFinder) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {}

func (v *PathFinder) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {}

func (v *PathFinder) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {}

func (v *PathFinder) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {}

func (v *PathFinder) ExprBinaryConcat(n *ast.ExprBinaryConcat) {}

func (v *PathFinder) ExprBinaryDiv(n *ast.ExprBinaryDiv) {}

func (v *PathFinder) ExprBinaryEqual(n *ast.ExprBinaryEqual) {}

func (v *PathFinder) ExprBinaryGreater(n *ast.ExprBinaryGreater) {}

func (v *PathFinder) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {}

func (v *PathFinder) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {}

func (v *PathFinder) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {}

func (v *PathFinder) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {}

func (v *PathFinder) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {}

func (v *PathFinder) ExprBinaryMinus(n *ast.ExprBinaryMinus) {}

func (v *PathFinder) ExprBinaryMod(n *ast.ExprBinaryMod) {}

func (v *PathFinder) ExprBinaryMul(n *ast.ExprBinaryMul) {}

func (v *PathFinder) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {}

func (v *PathFinder) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {}

func (v *PathFinder) ExprBinaryPlus(n *ast.ExprBinaryPlus) {}

func (v *PathFinder) ExprBinaryPow(n *ast.ExprBinaryPow) {}

func (v *PathFinder) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {}

func (v *PathFinder) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {}

func (v *PathFinder) ExprBinarySmaller(n *ast.ExprBinarySmaller) {}

func (v *PathFinder) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {}

func (v *PathFinder) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {}

func (v *PathFinder) ExprCastArray(n *ast.ExprCastArray) {}

func (v *PathFinder) ExprCastBool(n *ast.ExprCastBool) {}

func (v *PathFinder) ExprCastDouble(n *ast.ExprCastDouble) {}

func (v *PathFinder) ExprCastInt(n *ast.ExprCastInt) {}

func (v *PathFinder) ExprCastObject(n *ast.ExprCastObject) {}

func (v *PathFinder) ExprCastString(n *ast.ExprCastString) {}

func (v *PathFinder) ExprCastUnset(n *ast.ExprCastUnset) {}

func (v *PathFinder) ScalarDnumber(n *ast.ScalarDnumber) {}

func (v *PathFinder) ScalarEncapsed(n *ast.ScalarEncapsed) {}

func (v *PathFinder) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {}

func (v *PathFinder) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {}

func (v *PathFinder) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {}

func (v *PathFinder) ScalarHeredoc(n *ast.ScalarHeredoc) {}

func (v *PathFinder) ScalarLnumber(n *ast.ScalarLnumber) {}

func (v *PathFinder) ScalarMagicConstant(n *ast.ScalarMagicConstant) {}

func (v *PathFinder) ScalarString(n *ast.ScalarString) {}

func (v *PathFinder) NameName(n *ast.Name) {}

func (v *PathFinder) NameFullyQualified(n *ast.NameFullyQualified) {}

func (v *PathFinder) NameRelative(n *ast.NameRelative) {}

func (v *PathFinder) NameNamePart(n *ast.NamePart) {}
