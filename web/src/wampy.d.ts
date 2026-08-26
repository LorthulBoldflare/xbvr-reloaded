// wampy ships no TypeScript declarations; declare the minimal surface we use.
declare module 'wampy' {
  export interface WampyOptions {
    realm?: string
    onConnect?: () => void
    onClose?: () => void
    onError?: () => void
    onReconnect?: () => void
    onReconnectSuccess?: () => void
  }
  export class Wampy {
    constructor(url: string, options?: WampyOptions)
    subscribe(topic: string, handler: (args: any, kwargs?: any, details?: any) => void): void
    unsubscribe(topic: string): void
    disconnect(): void
  }
}
