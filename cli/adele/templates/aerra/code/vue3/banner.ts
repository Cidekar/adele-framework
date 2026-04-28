import { reactive } from 'vue'

// banner is a tiny global notification store the AlertBanner component
// subscribes to. Any code path can publish a message here — api.ts on a
// failed POST/GET, components on validation errors, etc. — and a single
// banner renders at the top of the viewport. Mirrors the OG aerra Jet
// flow where the layout's <body> renders {{.Error}} / {{.Flash}} above
// every page.
export type BannerKind = 'error' | 'flash' | null

export const banner = reactive<{
    kind: BannerKind
    message: string
}>({
    kind: null,
    message: '',
})

export function showError(message: string) {
    if (!message) {
        clearBanner()
        return
    }
    banner.kind = 'error'
    banner.message = message
}

export function showFlash(message: string) {
    if (!message) {
        clearBanner()
        return
    }
    banner.kind = 'flash'
    banner.message = message
}

export function clearBanner() {
    banner.kind = null
    banner.message = ''
}
