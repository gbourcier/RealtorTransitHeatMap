import { api, type Paginated } from './client'

export interface Schedule {
  id: string
  name: string
  cronExpr: string
  enabled: boolean
  createdAt: number
  updatedAt: number
}

export interface CreateScheduleRequest {
  name: string
  cronExpr: string
  enabled?: boolean
}

export interface UpdateScheduleRequest {
  name?: string
  cronExpr?: string
  enabled?: boolean
}

export interface ListSchedulesParams {
  limit?: number
  offset?: number
  enabled?: boolean
}

export async function listSchedules(params: ListSchedulesParams = {}): Promise<Paginated<Schedule>> {
  const { data } = await api.get<Paginated<Schedule>>('/api/schedules', { params })
  return data
}

export async function createSchedule(req: CreateScheduleRequest): Promise<Schedule> {
  const { data } = await api.post<Schedule>('/api/schedules', req)
  return data
}

export async function updateSchedule(id: string, req: UpdateScheduleRequest): Promise<Schedule> {
  const { data } = await api.patch<Schedule>(`/api/schedules/${id}`, req)
  return data
}

export async function deleteSchedule(id: string): Promise<void> {
  await api.delete(`/api/schedules/${id}`)
}
