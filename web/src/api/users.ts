import { api } from './client'
import type { User } from './auth'

export interface CreateUserRequest {
  username: string
  password: string
  role?: 'admin' | 'user'
}

export interface UpdateUserRequest {
  role?: 'admin' | 'user'
  isActive?: boolean
  password?: string
}

export async function listUsers(): Promise<User[]> {
  const { data } = await api.get<User[]>('/api/users')
  return data
}

export async function createUser(req: CreateUserRequest): Promise<User> {
  const { data } = await api.post<User>('/api/users', req)
  return data
}

export async function updateUser(id: string, req: UpdateUserRequest): Promise<User> {
  const { data } = await api.patch<User>(`/api/users/${id}`, req)
  return data
}

export async function deleteUser(id: string): Promise<void> {
  await api.delete(`/api/users/${id}`)
}
