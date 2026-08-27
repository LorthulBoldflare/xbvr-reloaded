import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../../api/client'
import { useOptionsState } from '../../../api/hooks'
import { useToastStore } from '../../../store/toasts'
import { Toggle } from '../../../components/Toggle'
import { ListEditor } from '../../../components/ListEditor'
import { SectionCard, Field, SaveButton, inputCls } from '../common'

// DLNA service settings.
export function DlnaSection() {
  const { data: state } = useOptionsState()
  const queryClient = useQueryClient()
  const toast = useToastStore.getState()

  const [enabled, setEnabled] = useState(true)
  const [name, setName] = useState('XBVR')
  const [image, setImage] = useState('default')
  const [allowedIp, setAllowedIp] = useState<string[]>([])

  useEffect(() => {
    const d = state?.config?.interfaces?.dlna
    if (d) {
      setEnabled(d.enabled)
      setName(d.serviceName)
      setImage(d.serviceImage)
      setAllowedIp(d.allowedIp ?? [])
    }
  }, [state?.config?.interfaces?.dlna])

  const save = useMutation({
    mutationFn: () => api.put('/options/interface/dlna', { enabled, name, image, allowedIp }),
    onSuccess: () => {
      toast.success('DLNA settings saved')
      queryClient.invalidateQueries({ queryKey: ['optionsState'] })
    }
  })

  const images = state?.currentState?.dlna?.images ?? []
  const recentIps = state?.currentState?.dlna?.recentIp ?? []

  return (
    <SectionCard title="DLNA" actions={<SaveButton onClick={() => save.mutate()} pending={save.isPending} />}>
      <div className="max-w-md space-y-3">
        <Toggle checked={enabled} onChange={setEnabled} label="Enabled" />
        <Field label="Service name">
          <input value={name} onChange={(e) => setName(e.target.value)} className={inputCls} />
        </Field>
        <Field label="Service icon">
          <select value={image} onChange={(e) => setImage(e.target.value)} className={inputCls}>
            <option value="default">default</option>
            {images.map((i) => (
              <option key={i} value={i}>
                {i}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Allowed IP addresses" hint="Empty = allow all">
          <ListEditor items={allowedIp} onChange={setAllowedIp} addLabel="Add IP" />
        </Field>
        {recentIps.length > 0 && (
          <Field label="Recently seen">
            <div className="flex flex-wrap gap-1">
              {recentIps.map((ip) => (
                <button
                  key={ip}
                  onClick={() => setAllowedIp((a) => (a.includes(ip) ? a : [...a, ip]))}
                  className="rounded-full bg-surface-3 px-2 py-0.5 font-mono text-xs hover:bg-accent-soft"
                  title="Add to allowed IPs"
                >
                  {ip}
                </button>
              ))}
            </div>
          </Field>
        )}
      </div>
    </SectionCard>
  )
}
