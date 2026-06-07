import { api } from './client'

export async function setDefaultFilter(defaultFilterId: string | null): Promise<void> {
  await api.patch('/api/preferences', { defaultFilterId })
}
