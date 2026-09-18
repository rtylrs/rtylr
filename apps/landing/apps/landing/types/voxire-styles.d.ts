// @voxire/components ships a stylesheet entry point with no type
// declaration, so the side-effect import in _app.tsx has nothing to
// resolve to. Declaring it here keeps the import honest without pretending
// the module has an API: it does not, it is CSS.
declare module '@voxire/components/styles';
