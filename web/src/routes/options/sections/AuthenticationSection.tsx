import { useEffect, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, ApiError } from '../../../api/client'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { SectionCard, Field, SaveButton, inputCls } from '../common'
import { buildDeovrPayload } from './deovrPayload'

// Authentication: player credentials (DeoVR/HereSphere) + MCP access token.
// Player credentials are saved through the existing deovr interface payload;
// the password round-trips as the redacted sentinel unless replaced.
export function AuthenticationSection() {
  const { data: state } = useOptionsState()
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()

  const [authEnabled, setAuthEnabled] = useState(false)
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showToken, setShowToken] = useState(false)

  useEffect(() => {
    const d = state?.config?.interfaces?.deovr
    if (d) {
      setAuthEnabled(d.auth_enabled)
      setUsername(d.username)
      setPassword(d.password) // redacted sentinel "***" when set
    }
  }, [state?.config?.interfaces?.deovr])

  // 404 when UI auth is disabled (no UI creds → no token can exist)
  const { data: tokenData, error } = useQuery({
    queryKey: ['mcpToken'],
    queryFn: () => api.get<{ token: string }>('/options/mcp-token', { toastOnError: false }),
    retry: false
  })
  const mcpUnavailable = error instanceof ApiError && error.status === 404

  const save = useMutation({
    mutationFn: () =>
      api.put(
        '/options/interface/deovr',
        buildDeovrPayload(state!.config, { auth_enabled: authEnabled, username, password })
      ),
    onSuccess: () => {
      toast.success('Authentication settings saved')
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
    }
  })

  return (
    <div className="space-y-4">
      <SectionCard title="Player credentials">
        <p className="mb-3 text-sm text-muted">
          Used by DeoVR/HereSphere players. A successful player login also unlocks this web UI in the headset
          browser (via a session cookie).
        </p>
        <div className="max-w-md space-y-3">
          <Toggle checked={authEnabled} onChange={setAuthEnabled} label="Authentication enabled" />
          <Field label="Username">
            <input value={username} onChange={(e) => setUsername(e.target.value)} className={inputCls} autoComplete="off" />
          </Field>
          <Field label="Password" hint="Shown as *** when a password is set — enter a new one to change it">
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className={inputCls}
              autoComplete="new-password"
            />
          </Field>
          <SaveButton onClick={() => save.mutate()} pending={save.isPending} />
        </div>
      </SectionCard>

      <SectionCard title="MCP access">
        {mcpUnavailable ? (
          <p className="text-sm text-muted">
            MCP is disabled. Set the <code className="font-mono">UI_USERNAME</code> and{' '}
            <code className="font-mono">UI_PASSWORD</code> environment variables to enable the MCP endpoint and UI
            authentication.
          </p>
        ) : (
          <>
            <p className="mb-3 text-sm text-muted">
              Bearer token for the MCP endpoint (<code className="font-mono">/mcp</code>). Derived from the UI
              credentials — it changes when they change.
            </p>
            <div className="flex max-w-xl items-center gap-2">
              <input
                readOnly
                type={showToken ? 'text' : 'password'}
                value={tokenData?.token ?? ''}
                className={`${inputCls} font-mono text-xs`}
              />
              <button onClick={() => setShowToken((s) => !s)} className="rounded-lg border border-line px-3 py-1.5 text-sm">
                {showToken ? 'Hide' : 'Reveal'}
              </button>
              <button
                onClick={() => {
                  navigator.clipboard.writeText(tokenData?.token ?? '')
                  toast.success('Token copied')
                }}
                className="rounded-lg border border-line px-3 py-1.5 text-sm"
              >
                Copy
              </button>
            </div>
          </>
        )}
      </SectionCard>
    </div>
  )
}
