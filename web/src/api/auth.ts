import { api } from './client'

export interface User {
  id: string
  username: string
  role: 'admin' | 'user'
  isActive: boolean
  createdAt: number
  updatedAt: number
}

export async function login(username: string, password: string): Promise<User> {
  const { data } = await api.post<User>('/api/auth/login', { username, password })
  return data
}

export async function logout(): Promise<void> {
  await api.post('/api/auth/logout')
}

export async function me(): Promise<User> {
  const { data } = await api.get<User>('/api/auth/me')
  return data
}
