/** Lightweight event emitter for the auth app. */
type Listener<T> = (payload: T) => void;
export class EventEmitter<EventMap extends Record<string, unknown>> {
  private readonly listeners: Partial<{ [K in keyof EventMap]: Set<Listener<EventMap[K]>> }> = {};
  on<K extends keyof EventMap>(event: K, listener: Listener<EventMap[K]>): () => void {
    if (!this.listeners[event]) this.listeners[event] = new Set();
    this.listeners[event]!.add(listener);
    return () => this.off(event, listener);
  }
  off<K extends keyof EventMap>(event: K, listener: Listener<EventMap[K]>): void {
    this.listeners[event]?.delete(listener);
  }
  emit<K extends keyof EventMap>(event: K, payload: EventMap[K]): void {
    this.listeners[event]?.forEach((fn) => fn(payload));
  }
  clear(): void {
    for (const key of Object.keys(this.listeners) as (keyof EventMap)[]) delete this.listeners[key];
  }
}
