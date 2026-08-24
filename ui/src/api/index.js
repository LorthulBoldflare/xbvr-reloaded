import ky from 'ky'
import { ToastProgrammatic as Toast } from 'buefy'

// Shared API client for the XBVR backend.
// - prefixUrl '/api' so call sites use resource paths only
// - 60s default timeout; long-running endpoints (scene/list, bundle
//   backup/restore, filters) opt out with their own per-call budgets
//   matching the historical behavior
// - failed requests surface a toast and carry the response body in the
//   error message
const kyInstance = ky.create({
  prefixUrl: '/api',
  timeout: 60000,
  hooks: {
    beforeError: [
      async error => {
        const { response } = error
        if (response) {
          try {
            const body = await response.clone().text()
            if (body) {
              error.message = `${response.status} ${body.slice(0, 200)}`
            }
          } catch (_) { /* keep original message */ }
          Toast.open({
            message: `Request failed: ${error.message}`,
            type: 'is-danger',
            duration: 5000
          })
        }
        return error
      }
    ]
  }
})

// ky rejects inputs starting with '/' when prefixUrl is set, but most call
// sites migrated from absolute '/api/...' URLs still pass a leading slash —
// normalize here instead of touching every call site. ky instance methods
// are closure-bound (no `this`), so delegating like this is safe.
const stripLeadingSlash = input => (typeof input === 'string' ? input.replace(/^\/+/, '') : input)

const api = {}
for (const method of ['get', 'post', 'put', 'patch', 'delete', 'head']) {
  api[method] = (input, options) => kyInstance[method](stripLeadingSlash(input), options)
}
api.extend = kyInstance.extend

export default api
