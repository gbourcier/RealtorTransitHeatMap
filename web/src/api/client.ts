import axios from 'axios'

export const api = axios.create({
  baseURL: '/',
  headers: { 'Content-Type': 'application/json' },
})

export interface Paginated<T> {
  items: T[]
  total: number
  limit: number
  offset: number
}
