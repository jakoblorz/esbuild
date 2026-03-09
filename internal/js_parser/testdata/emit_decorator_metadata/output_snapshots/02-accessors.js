function dec(...args) {
}
class Accessors {
  get typedGet() {
    return "";
  }
  get untypedGet() {
    return "";
  }
  set typedSet(value) {
  }
  set untypedSet(value) {
  }
}
__decorateClass([
  dec,
  __legacyMetadata("design:type", String),
  __legacyMetadata("design:paramtypes", [])
], Accessors.prototype, "typedGet", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object),
  __legacyMetadata("design:paramtypes", [])
], Accessors.prototype, "untypedGet", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", String),
  __legacyMetadata("design:paramtypes", [
    String
  ])
], Accessors.prototype, "typedSet", 1);
__decorateClass([
  dec,
  __legacyMetadata("design:type", Object),
  __legacyMetadata("design:paramtypes", [
    Object
  ])
], Accessors.prototype, "untypedSet", 1);
