// ----- 15-type-only-namespace.index.ts -----
class Main {
}
__decorateClass([
  decorator,
  __legacyMetadata("design:type", Function)
], Main.prototype, "field", 2);

// ----- 15-type-only-namespace.services.ts -----
export var Services;
((Services) => {
  class Service {
  }
  Services.Service = Service;
})(Services || (Services = {}));
