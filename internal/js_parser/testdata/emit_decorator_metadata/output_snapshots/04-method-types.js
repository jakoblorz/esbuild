function dec(...args) {
}
class Ref {
}
var Kind = /* @__PURE__ */ ((Kind) => {
  Kind[Kind["A"] = 0] = "A";
  return Kind;
})(Kind || {});
class MethodTypes {
  rich(a, b, c, d, e, f, g) {
    return "";
  }
  implicit(value) {
  }
}
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    Object,
    String,
    String,
    Object,
    Number,
    Array,
    Function
  ]),
  __legacyMetadata("design:returntype", String)
], MethodTypes.prototype, "rich", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Function),
  __legacyMetadata("design:paramtypes", [
    Object
  ]),
  __legacyMetadata("design:returntype", void 0)
], MethodTypes.prototype, "implicit", 1);
