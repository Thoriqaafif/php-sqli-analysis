package SQLiDetector

import "github.com/z7zmey/php-parser/pkg/ast"

type SQLiDetector struct{}

func (v *SQLiDetector) Root(n *ast.Root) {}

func (v *SQLiDetector) Nullable(n *ast.Nullable) {}

func (v *SQLiDetector) Parameter(n *ast.Parameter) {}

func (v *SQLiDetector) Identifier(n *ast.Identifier) {}

func (v *SQLiDetector) Argument(n *ast.Argument) {}

func (v *SQLiDetector) Attribute(n *ast.Attribute) {}

func (v *SQLiDetector) AttributeGroup(n *ast.AttributeGroup) {}

func (v *SQLiDetector) StmtBreak(n *ast.StmtBreak) {}

func (v *SQLiDetector) StmtCase(n *ast.StmtCase) {}

func (v *SQLiDetector) StmtCatch(n *ast.StmtCatch) {}

func (v *SQLiDetector) StmtClass(n *ast.StmtClass) {}

func (v *SQLiDetector) StmtClassConstList(n *ast.StmtClassConstList) {}

func (v *SQLiDetector) StmtClassMethod(n *ast.StmtClassMethod) {}

func (v *SQLiDetector) StmtConstList(n *ast.StmtConstList) {}

func (v *SQLiDetector) StmtConstant(n *ast.StmtConstant) {}

func (v *SQLiDetector) StmtContinue(n *ast.StmtContinue) {}

func (v *SQLiDetector) StmtDeclare(n *ast.StmtDeclare) {}

func (v *SQLiDetector) StmtDefault(n *ast.StmtDefault) {}

func (v *SQLiDetector) StmtDo(n *ast.StmtDo) {}

func (v *SQLiDetector) StmtEcho(n *ast.StmtEcho) {}

func (v *SQLiDetector) StmtElse(n *ast.StmtElse) {}

func (v *SQLiDetector) StmtElseIf(n *ast.StmtElseIf) {}

func (v *SQLiDetector) StmtExpression(n *ast.StmtExpression) {}

func (v *SQLiDetector) StmtFinally(n *ast.StmtFinally) {}

func (v *SQLiDetector) StmtFor(n *ast.StmtFor) {}

func (v *SQLiDetector) StmtForeach(n *ast.StmtForeach) {}

func (v *SQLiDetector) StmtFunction(n *ast.StmtFunction) {}

func (v *SQLiDetector) StmtGlobal(n *ast.StmtGlobal) {}

func (v *SQLiDetector) StmtGoto(n *ast.StmtGoto) {}

func (v *SQLiDetector) StmtHaltCompiler(n *ast.StmtHaltCompiler) {}

func (v *SQLiDetector) StmtIf(n *ast.StmtIf) {}

func (v *SQLiDetector) StmtInlineHtml(n *ast.StmtInlineHtml) {}

func (v *SQLiDetector) StmtInterface(n *ast.StmtInterface) {}

func (v *SQLiDetector) StmtLabel(n *ast.StmtLabel) {}

func (v *SQLiDetector) StmtNamespace(n *ast.StmtNamespace) {}

func (v *SQLiDetector) StmtNop(n *ast.StmtNop) {}

func (v *SQLiDetector) StmtProperty(n *ast.StmtProperty) {}

func (v *SQLiDetector) StmtPropertyList(n *ast.StmtPropertyList) {}

func (v *SQLiDetector) StmtReturn(n *ast.StmtReturn) {}

func (v *SQLiDetector) StmtStatic(n *ast.StmtStatic) {}

func (v *SQLiDetector) StmtStaticVar(n *ast.StmtStaticVar) {}

func (v *SQLiDetector) StmtStmtList(n *ast.StmtStmtList) {}

func (v *SQLiDetector) StmtSwitch(n *ast.StmtSwitch) {}

func (v *SQLiDetector) StmtThrow(n *ast.StmtThrow) {}

func (v *SQLiDetector) StmtTrait(n *ast.StmtTrait) {}

func (v *SQLiDetector) StmtTraitUse(n *ast.StmtTraitUse) {}

func (v *SQLiDetector) StmtTraitUseAlias(n *ast.StmtTraitUseAlias) {}

func (v *SQLiDetector) StmtTraitUsePrecedence(n *ast.StmtTraitUsePrecedence) {}

func (v *SQLiDetector) StmtTry(n *ast.StmtTry) {}

func (v *SQLiDetector) StmtUnset(n *ast.StmtUnset) {}

func (v *SQLiDetector) StmtUse(n *ast.StmtUseList) {}

func (v *SQLiDetector) StmtGroupUse(n *ast.StmtGroupUseList) {}

func (v *SQLiDetector) StmtUseDeclaration(n *ast.StmtUse) {}

func (v *SQLiDetector) StmtWhile(n *ast.StmtWhile) {}

func (v *SQLiDetector) ExprArray(n *ast.ExprArray) {}

func (v *SQLiDetector) ExprArrayDimFetch(n *ast.ExprArrayDimFetch) {}

func (v *SQLiDetector) ExprArrayItem(n *ast.ExprArrayItem) {}

func (v *SQLiDetector) ExprArrowFunction(n *ast.ExprArrowFunction) {}

func (v *SQLiDetector) ExprBrackets(n *ast.ExprBrackets) {}

func (v *SQLiDetector) ExprBitwiseNot(n *ast.ExprBitwiseNot) {}

func (v *SQLiDetector) ExprBooleanNot(n *ast.ExprBooleanNot) {}

func (v *SQLiDetector) ExprClassConstFetch(n *ast.ExprClassConstFetch) {}

func (v *SQLiDetector) ExprClone(n *ast.ExprClone) {}

func (v *SQLiDetector) ExprClosure(n *ast.ExprClosure) {}

func (v *SQLiDetector) ExprClosureUse(n *ast.ExprClosureUse) {}

func (v *SQLiDetector) ExprConstFetch(n *ast.ExprConstFetch) {}

func (v *SQLiDetector) ExprEmpty(n *ast.ExprEmpty) {}

func (v *SQLiDetector) ExprErrorSuppress(n *ast.ExprErrorSuppress) {}

func (v *SQLiDetector) ExprEval(n *ast.ExprEval) {}

func (v *SQLiDetector) ExprExit(n *ast.ExprExit) {}

func (v *SQLiDetector) ExprFunctionCall(n *ast.ExprFunctionCall) {}

func (v *SQLiDetector) ExprInclude(n *ast.ExprInclude) {}

func (v *SQLiDetector) ExprIncludeOnce(n *ast.ExprIncludeOnce) {}

func (v *SQLiDetector) ExprInstanceOf(n *ast.ExprInstanceOf) {}

func (v *SQLiDetector) ExprIsset(n *ast.ExprIsset) {}

func (v *SQLiDetector) ExprList(n *ast.ExprList) {}

func (v *SQLiDetector) ExprMethodCall(n *ast.ExprMethodCall) {}

func (v *SQLiDetector) ExprNew(n *ast.ExprNew) {}

func (v *SQLiDetector) ExprPostDec(n *ast.ExprPostDec) {}

func (v *SQLiDetector) ExprPostInc(n *ast.ExprPostInc) {}

func (v *SQLiDetector) ExprPreDec(n *ast.ExprPreDec) {}

func (v *SQLiDetector) ExprPreInc(n *ast.ExprPreInc) {}

func (v *SQLiDetector) ExprPrint(n *ast.ExprPrint) {}

func (v *SQLiDetector) ExprPropertyFetch(n *ast.ExprPropertyFetch) {}

func (v *SQLiDetector) ExprRequire(n *ast.ExprRequire) {}

func (v *SQLiDetector) ExprRequireOnce(n *ast.ExprRequireOnce) {}

func (v *SQLiDetector) ExprShellExec(n *ast.ExprShellExec) {}

func (v *SQLiDetector) ExprStaticCall(n *ast.ExprStaticCall) {}

func (v *SQLiDetector) ExprStaticPropertyFetch(n *ast.ExprStaticPropertyFetch) {}

func (v *SQLiDetector) ExprTernary(n *ast.ExprTernary) {}

func (v *SQLiDetector) ExprUnaryMinus(n *ast.ExprUnaryMinus) {}

func (v *SQLiDetector) ExprUnaryPlus(n *ast.ExprUnaryPlus) {}

func (v *SQLiDetector) ExprVariable(n *ast.ExprVariable) {}

func (v *SQLiDetector) ExprYield(n *ast.ExprYield) {}

func (v *SQLiDetector) ExprYieldFrom(n *ast.ExprYieldFrom) {}

func (v *SQLiDetector) ExprAssign(n *ast.ExprAssign) {}

func (v *SQLiDetector) ExprAssignReference(n *ast.ExprAssignReference) {}

func (v *SQLiDetector) ExprAssignBitwiseAnd(n *ast.ExprAssignBitwiseAnd) {}

func (v *SQLiDetector) ExprAssignBitwiseOr(n *ast.ExprAssignBitwiseOr) {}

func (v *SQLiDetector) ExprAssignBitwiseXor(n *ast.ExprAssignBitwiseXor) {}

func (v *SQLiDetector) ExprAssignCoalesce(n *ast.ExprAssignCoalesce) {}

func (v *SQLiDetector) ExprAssignConcat(n *ast.ExprAssignConcat) {}

func (v *SQLiDetector) ExprAssignDiv(n *ast.ExprAssignDiv) {}

func (v *SQLiDetector) ExprAssignMinus(n *ast.ExprAssignMinus) {}

func (v *SQLiDetector) ExprAssignMod(n *ast.ExprAssignMod) {}

func (v *SQLiDetector) ExprAssignMul(n *ast.ExprAssignMul) {}

func (v *SQLiDetector) ExprAssignPlus(n *ast.ExprAssignPlus) {}

func (v *SQLiDetector) ExprAssignPow(n *ast.ExprAssignPow) {}

func (v *SQLiDetector) ExprAssignShiftLeft(n *ast.ExprAssignShiftLeft) {}

func (v *SQLiDetector) ExprAssignShiftRight(n *ast.ExprAssignShiftRight) {}

func (v *SQLiDetector) ExprBinaryBitwiseAnd(n *ast.ExprBinaryBitwiseAnd) {}

func (v *SQLiDetector) ExprBinaryBitwiseOr(n *ast.ExprBinaryBitwiseOr) {}

func (v *SQLiDetector) ExprBinaryBitwiseXor(n *ast.ExprBinaryBitwiseXor) {}

func (v *SQLiDetector) ExprBinaryBooleanAnd(n *ast.ExprBinaryBooleanAnd) {}

func (v *SQLiDetector) ExprBinaryBooleanOr(n *ast.ExprBinaryBooleanOr) {}

func (v *SQLiDetector) ExprBinaryCoalesce(n *ast.ExprBinaryCoalesce) {}

func (v *SQLiDetector) ExprBinaryConcat(n *ast.ExprBinaryConcat) {}

func (v *SQLiDetector) ExprBinaryDiv(n *ast.ExprBinaryDiv) {}

func (v *SQLiDetector) ExprBinaryEqual(n *ast.ExprBinaryEqual) {}

func (v *SQLiDetector) ExprBinaryGreater(n *ast.ExprBinaryGreater) {}

func (v *SQLiDetector) ExprBinaryGreaterOrEqual(n *ast.ExprBinaryGreaterOrEqual) {}

func (v *SQLiDetector) ExprBinaryIdentical(n *ast.ExprBinaryIdentical) {}

func (v *SQLiDetector) ExprBinaryLogicalAnd(n *ast.ExprBinaryLogicalAnd) {}

func (v *SQLiDetector) ExprBinaryLogicalOr(n *ast.ExprBinaryLogicalOr) {}

func (v *SQLiDetector) ExprBinaryLogicalXor(n *ast.ExprBinaryLogicalXor) {}

func (v *SQLiDetector) ExprBinaryMinus(n *ast.ExprBinaryMinus) {}

func (v *SQLiDetector) ExprBinaryMod(n *ast.ExprBinaryMod) {}

func (v *SQLiDetector) ExprBinaryMul(n *ast.ExprBinaryMul) {}

func (v *SQLiDetector) ExprBinaryNotEqual(n *ast.ExprBinaryNotEqual) {}

func (v *SQLiDetector) ExprBinaryNotIdentical(n *ast.ExprBinaryNotIdentical) {}

func (v *SQLiDetector) ExprBinaryPlus(n *ast.ExprBinaryPlus) {}

func (v *SQLiDetector) ExprBinaryPow(n *ast.ExprBinaryPow) {}

func (v *SQLiDetector) ExprBinaryShiftLeft(n *ast.ExprBinaryShiftLeft) {}

func (v *SQLiDetector) ExprBinaryShiftRight(n *ast.ExprBinaryShiftRight) {}

func (v *SQLiDetector) ExprBinarySmaller(n *ast.ExprBinarySmaller) {}

func (v *SQLiDetector) ExprBinarySmallerOrEqual(n *ast.ExprBinarySmallerOrEqual) {}

func (v *SQLiDetector) ExprBinarySpaceship(n *ast.ExprBinarySpaceship) {}

func (v *SQLiDetector) ExprCastArray(n *ast.ExprCastArray) {}

func (v *SQLiDetector) ExprCastBool(n *ast.ExprCastBool) {}

func (v *SQLiDetector) ExprCastDouble(n *ast.ExprCastDouble) {}

func (v *SQLiDetector) ExprCastInt(n *ast.ExprCastInt) {}

func (v *SQLiDetector) ExprCastObject(n *ast.ExprCastObject) {}

func (v *SQLiDetector) ExprCastString(n *ast.ExprCastString) {}

func (v *SQLiDetector) ExprCastUnset(n *ast.ExprCastUnset) {}

func (v *SQLiDetector) ScalarDnumber(n *ast.ScalarDnumber) {}

func (v *SQLiDetector) ScalarEncapsed(n *ast.ScalarEncapsed) {}

func (v *SQLiDetector) ScalarEncapsedStringPart(n *ast.ScalarEncapsedStringPart) {}

func (v *SQLiDetector) ScalarEncapsedStringVar(n *ast.ScalarEncapsedStringVar) {}

func (v *SQLiDetector) ScalarEncapsedStringBrackets(n *ast.ScalarEncapsedStringBrackets) {}

func (v *SQLiDetector) ScalarHeredoc(n *ast.ScalarHeredoc) {}

func (v *SQLiDetector) ScalarLnumber(n *ast.ScalarLnumber) {}

func (v *SQLiDetector) ScalarMagicConstant(n *ast.ScalarMagicConstant) {}

func (v *SQLiDetector) ScalarString(n *ast.ScalarString) {}

func (v *SQLiDetector) NameName(n *ast.Name) {}

func (v *SQLiDetector) NameFullyQualified(n *ast.NameFullyQualified) {}

func (v *SQLiDetector) NameRelative(n *ast.NameRelative) {}

func (v *SQLiDetector) NameNamePart(n *ast.NamePart) {}
