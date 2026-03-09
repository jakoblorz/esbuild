// ----- 13-type-only-import.component.ts -----
let MyComponent = class {
  constructor(service) {
    this.service = service;
  }
  method(x) {
  }
};
__decorateClass([
  decorator,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    Object
  ]),
  __legacyMetadata("design:returntype", void 0)
], MyComponent.prototype, "method", 1);
MyComponent = __decorateClass([
  decorator,
  __legacyMetadata("design:paramtypes", [
    Function
  ])
], MyComponent);

// ----- 13-type-only-import.service.ts -----
export class Service {
}
