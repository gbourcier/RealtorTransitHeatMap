import { api, type Paginated } from './client'

export type ScrapeStatus = 'running' | 'success' | 'error'

export interface ScrapeRun {
  id: string
  source: string
  status: ScrapeStatus
  startedAt: number
  completedAt?: number | null
  totalCount?: number | null
  newCount?: number | null
  staleCount?: number | null
  errorKind?: string
  errorMessage?: string
  scheduleId?: string | null
  scheduleName?: string | null
}

export interface ListRunsParams {
  limit?: number
  offset?: number
  status?: ScrapeStatus
  scheduleId?: string
}

export async function listRuns(params: ListRunsParams = {}): Promise<Paginated<ScrapeRun>> {
  const { data } = await api.get<Paginated<ScrapeRun>>('/api/scrapes', { params })
  return data
}

export async function getRun(id: string): Promise<ScrapeRun> {
  const { data } = await api.get<ScrapeRun>(`/api/scrapes/${id}`)
  return data
}
