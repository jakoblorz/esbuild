function dec(...args) {
}
let Basic = class {
  constructor(a) {
  }
  method(x) {
    return true;
  }
};
__decorateClass([
  dec,
  __legacyMetadata("design:type", Number)
], Basic.prototype, "field", 2);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    String
  ]),
  __legacyMetadata("design:returntype", Boolean)
], Basic.prototype, "method", 1);
Basic = __decorateClass([
  dec,
  __legacyMetadata("design:paramtypes", [
    String
  ])
], Basic);
