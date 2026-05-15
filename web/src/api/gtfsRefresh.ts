import { api, type Paginated } from './client'

export type RefreshStatus = 'running' | 'success' | 'error'

export interface RefreshRun {
  id: string
  status: RefreshStatus
  startedAt: number
  completedAt?: number | null
  stopsWritten?: number | null
  errorMessage?: string
  scheduleId?: string | null
}

export interface ListRefreshRunsParams {
  limit?: number
  offset?: number
  status?: RefreshStatus
  scheduleId?: string
}

export async function listRefreshRuns(
  params: ListRefreshRunsParams = {},
): Promise<Paginated<RefreshRun>> {
  const { data } = await api.get<Paginated<RefreshRun>>('/api/gtfs-refresh-runs', { params })
  return data
}

export async function getRefreshRun(id: string): Promise<RefreshRun> {
  const { data } = await api.get<RefreshRun>(`/api/gtfs-refresh-runs/${id}`)
  return data
}

export async function startRefresh(): Promise<{ runId: string }> {
  const { data } = await api.post<{ runId: string }>('/api/gtfs-refresh-runs')
  return data
}
