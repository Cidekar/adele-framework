import { showError, showFlash, clearBanner } from './banner'

type ApiResponse<T = unknown> = {
    ok: boolean
    status?: number
    redirect?: string
    message?: string
    errors?: Record<string, string>
    user?: T
} & Record<string, unknown>

function csrfToken(): string {
    const meta = document.querySelector('meta[name="csrf-token"]')
    return meta?.getAttribute('content') ?? ''
}

// On a 401 to a private endpoint, route the SPA to /login so the user sees the
// login page instead of an empty dashboard. AppLayout already redirects on
// !res.ok, but doing it here covers any caller that forgets.
function handleUnauthorized(status: number) {
    if (status === 401 && typeof window !== 'undefined' && window.location.pathname !== '/login') {
        window.location.href = '/login'
    }
}

// publishBanner promotes a server-supplied message to the global AlertBanner.
// Field-level validation errors stay inline next to their inputs; only the
// top-level `message` (auth failures, generic 500s, success flashes) flows
// here. Mirrors the OG aerra Jet flow where Session.Put("error", ...) renders
// a full-width banner above the page via application.jet.
function publishBanner<T>(data: ApiResponse<T>) {
    if (data.ok && data.message) {
        showFlash(data.message)
        return
    }
    if (!data.ok && data.message) {
        showError(data.message)
        return
    }
    // Successful response without a message — clear any stale banner so it
    // doesn't carry over from a prior failed submit.
    if (data.ok) clearBanner()
}

export async function apiPost<T = unknown>(
    path: string,
    fields: Record<string, string>,
): Promise<ApiResponse<T>> {
    const body = new URLSearchParams()
    body.append('csrf_token', csrfToken())
    for (const [k, v] of Object.entries(fields)) body.append(k, v)

    const res = await fetch(path, {
        method: 'POST',
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body,
        credentials: 'include',
    })

    let data: ApiResponse<T>
    try {
        data = (await res.json()) as ApiResponse<T>
    } catch {
        data = { ok: false, message: `Server returned ${res.status}` }
    }
    data.status = res.status
    handleUnauthorized(res.status)
    publishBanner(data)
    return data
}

export async function apiGet<T = unknown>(path: string): Promise<ApiResponse<T>> {
    const res = await fetch(path, {
        method: 'GET',
        headers: { 'Accept': 'application/json' },
        credentials: 'include',
    })
    let data: ApiResponse<T>
    try {
        data = (await res.json()) as ApiResponse<T>
    } catch {
        data = { ok: false, message: `Server returned ${res.status}` }
    }
    data.status = res.status
    handleUnauthorized(res.status)
    // GET responses generally don't surface a message worth banner-ing
    // (Profile/Dashboard returning user data is silent on success). Only
    // promote on failure.
    if (!data.ok) publishBanner(data)
    return data
}
