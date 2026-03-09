function dec(...args) {
}
class Ref {
}
class InferredField {
  constructor() {
    this.inferred = new Ref();
    this.text = "x";
    this.num = 1;
    this.bool = true;
  }
}
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object)
], InferredField.prototype, "inferred", 2);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object)
], InferredField.prototype, "text", 2);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object)
], InferredField.prototype, "num", 2);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object)
], InferredField.prototype, "bool", 2);
