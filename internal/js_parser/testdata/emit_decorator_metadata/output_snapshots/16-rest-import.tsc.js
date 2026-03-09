// ----- 16-rest-import.aux.ts -----
export class SomeClass {
}

// ----- 16-rest-import.main.ts -----
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
import { SomeClass } from './16-rest-import.aux';
let ClassA = class ClassA {
    constructor(...init) { }
    foo(...args) { }
};
__decorate([
    annotation1,
    __metadata("design:type", Function),
    __metadata("design:paramtypes", [SomeClass]),
    __metadata("design:returntype", void 0)
], ClassA.prototype, "foo", null);
ClassA = __decorate([
    annotation,
    __metadata("design:paramtypes", [SomeClass])
], ClassA);
