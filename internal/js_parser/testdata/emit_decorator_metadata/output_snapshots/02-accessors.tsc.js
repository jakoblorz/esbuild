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
class Accessors {
    get typedGet() {
        return '';
    }
    get untypedGet() {
        return '';
    }
    set typedSet(value) { }
    set untypedSet(value) { }
}
__decorate([
    dec,
    __metadata("design:type", String),
    __metadata("design:paramtypes", [])
], Accessors.prototype, "typedGet", null);
__decorate([
    dec,
    __metadata("design:type", Object),
    __metadata("design:paramtypes", [])
], Accessors.prototype, "untypedGet", null);
__decorate([
    dec,
    __metadata("design:type", String),
    __metadata("design:paramtypes", [String])
], Accessors.prototype, "typedSet", null);
__decorate([
    dec,
    __metadata("design:type", Object),
    __metadata("design:paramtypes", [Object])
], Accessors.prototype, "untypedSet", null);
