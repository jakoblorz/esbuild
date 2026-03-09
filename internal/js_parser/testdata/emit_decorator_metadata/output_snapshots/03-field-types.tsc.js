var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
function dec(...args) { }
class Ref {
}
var Kind;
(function (Kind) {
    Kind[Kind["A"] = 0] = "A";
})(Kind || (Kind = {}));
class FieldTypes {
}
__decorate([
    dec,
    __metadata("design:type", Object)
], FieldTypes.prototype, "anyType", void 0);
__decorate([
    dec,
    __metadata("design:type", Object)
], FieldTypes.prototype, "unknownType", void 0);
__decorate([
    dec,
    __metadata("design:type", Object)
], FieldTypes.prototype, "objectType", void 0);
__decorate([
    dec,
    __metadata("design:type", String)
], FieldTypes.prototype, "stringType", void 0);
__decorate([
    dec,
    __metadata("design:type", Number)
], FieldTypes.prototype, "numberType", void 0);
__decorate([
    dec,
    __metadata("design:type", Boolean)
], FieldTypes.prototype, "booleanType", void 0);
__decorate([
    dec,
    __metadata("design:type", Symbol)
], FieldTypes.prototype, "symbolType", void 0);
__decorate([
    dec,
    __metadata("design:type", BigInt)
], FieldTypes.prototype, "bigintType", void 0);
__decorate([
    dec,
    __metadata("design:type", String)
], FieldTypes.prototype, "literalType", void 0);
__decorate([
    dec,
    __metadata("design:type", Array)
], FieldTypes.prototype, "arrayType", void 0);
__decorate([
    dec,
    __metadata("design:type", Array)
], FieldTypes.prototype, "tupleType", void 0);
__decorate([
    dec,
    __metadata("design:type", Object)
], FieldTypes.prototype, "unionType", void 0);
__decorate([
    dec,
    __metadata("design:type", Function)
], FieldTypes.prototype, "fnType", void 0);
__decorate([
    dec,
    __metadata("design:type", Ref)
], FieldTypes.prototype, "classType", void 0);
__decorate([
    dec,
    __metadata("design:type", Object)
], FieldTypes.prototype, "ifaceType", void 0);
__decorate([
    dec,
    __metadata("design:type", String)
], FieldTypes.prototype, "aliasType", void 0);
__decorate([
    dec,
    __metadata("design:type", Number)
], FieldTypes.prototype, "enumType", void 0);
__decorate([
    dec,
    __metadata("design:type", Promise)
], FieldTypes.prototype, "promiseType", void 0);
