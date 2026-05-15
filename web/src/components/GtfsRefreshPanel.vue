<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import {
  listRefreshRuns,
  startRefresh,
  type RefreshRun,
  type RefreshStatus,
} from '../api/gtfsRefresh'

const items = ref<RefreshRun[]>([])
const total = ref(0)
const limit = ref(20)
const offset = ref(0)
const loading = ref(false)
const triggering = ref(false)
const error = ref<string | null>(null)

let pollTimer: number | null = null

const hasRunning = computed(() => items.value.some((r) => r.status === 'running'))

async function refresh() {
  loading.value = true
  error.value = null
  try {
    const res = await listRefreshRuns({ limit: limit.value, offset: offset.value })
    items.value = res.items
    total.value = res.total
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? e?.message ?? 'failed to load runs'
  } finally {
    loading.value = false
  }
}

async function onTrigger() {
  triggering.value = true
  error.value = null
  try {
    await startRefresh()
    await refresh()
  } catch (e: any) {
    if (e?.response?.status === 409) {
      error.value = 'A refresh is already in progress.'
    } else {
      error.value = e?.response?.data?.error ?? e?.message ?? 'trigger failed'
    }
  } finally {
    triggering.value = false
  }
}

function statusColor(s: RefreshStatus): string {
  switch (s) {
    case 'running':
      return 'info'
    case 'success':
      return 'success'
    case 'error':
      return 'error'
  }
}

function formatTime(unix?: number | null): string {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleString()
}

function formatTimeShort(unix?: number | null): string {
  if (!unix) return '—'
  const date = new Date(unix * 1000)
  const now = new Date()
  const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate())
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const daysAgo = Math.round(
    (startOfToday.getTime() - startOfDate.getTime()) / 86400000,
  )
  const time = date.toLocaleTimeString([], {
    hour: 'numeric',
    minute: '2-digit',
  })
  if (daysAgo === 0) return `today ${time}`
  if (daysAgo === 1) return `yesterday ${time}`
  const dd = String(date.getDate()).padStart(2, '0')
  const mm = String(date.getMonth() + 1).padStart(2, '0')
  return `${dd}/${mm} ${time}`
}

function durationLabel(r: RefreshRun): string {
  const end = r.completedAt ?? Math.floor(Date.now() / 1000)
  const secs = Math.max(0, end - r.startedAt)
  if (secs < 60) return `${secs}s`
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m}m ${s}s`
}

onMounted(async () => {
  await refresh()
  pollTimer = window.setInterval(() => {
    if (hasRunning.value) refresh()
  }, 2000)
})

onUnmounted(() => {
  if (pollTimer !== null) window.clearInterval(pollTimer)
})
</script>

<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <span>GTFS Refresh Jobs</span>
      <v-spacer />
      <v-btn color="primary" prepend-icon="mdi-refresh" :loading="triggering" :disabled="hasRunning"
        class="d-none d-sm-inline-flex" @click="onTrigger">
        Refresh now
      </v-btn>
      <v-btn color="primary" icon="mdi-refresh" :loading="triggering" :disabled="hasRunning" size="small"
        class="d-inline-flex d-sm-none" @click="onTrigger" />
    </v-card-title>
    <v-divider />
    <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{ error }}</v-alert>
    <v-card-text v-if="loading && items.length === 0" class="text-center py-8">
      <v-progress-circular indeterminate />
    </v-card-text>
    <template v-else-if="items.length > 0">
      <v-table density="comfortable" class="d-none d-md-table">
        <thead>
          <tr>
            <th>Status</th>
            <th>Started</th>
            <th>Duration</th>
            <th class="text-right">Stops written</th>
            <th>Trigger</th>
            <th>Error</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="r in items" :key="r.id">
            <td>
              <v-chip size="small" :color="statusColor(r.status)" variant="tonal">
                {{ r.status }}
              </v-chip>
            </td>
            <td>{{ formatTime(r.startedAt) }}</td>
            <td>{{ durationLabel(r) }}</td>
            <td class="text-right">{{ r.stopsWritten != null ? r.stopsWritten.toLocaleString() : '—' }}</td>
            <td>
              <v-chip v-if="r.scheduleId" size="x-small" variant="outlined">scheduled</v-chip>
              <v-chip v-else size="x-small" variant="outlined" color="grey">manual</v-chip>
            </td>
            <td class="text-caption">
              <span v-if="r.errorMessage">{{ r.errorMessage }}</span>
            </td>
          </tr>
        </tbody>
      </v-table>

      <div class="d-md-none refresh-cards">
        <div v-for="r in items" :key="`m-${r.id}`" class="refresh-card">
          <div class="refresh-card__top">
            <v-chip size="small" :color="statusColor(r.status)" variant="tonal">
              {{ r.status }}
            </v-chip>
            <span class="refresh-card__when">{{ formatTimeShort(r.startedAt) }}</span>
          </div>
          <div class="refresh-card__stats">
            <span class="refresh-card__stat">
              <span class="refresh-card__stat-label">Duration</span>
              <span class="refresh-card__stat-value">{{ durationLabel(r) }}</span>
            </span>
            <span class="refresh-card__stat">
              <span class="refresh-card__stat-label">Stops</span>
              <span class="refresh-card__stat-value">
                {{ r.stopsWritten != null ? r.stopsWritten.toLocaleString() : '—' }}
              </span>
            </span>
          </div>
          <div class="refresh-card__bottom">
            <v-chip v-if="r.scheduleId" size="x-small" variant="outlined">scheduled</v-chip>
            <v-chip v-else size="x-small" variant="outlined" color="grey">manual</v-chip>
            <span v-if="r.errorMessage" class="refresh-card__error">{{ r.errorMessage }}</span>
          </div>
        </div>
      </div>
    </template>
    <v-card-text v-else class="text-medium-emphasis text-center py-8">
      No runs yet. Click "Refresh now" to start one.
    </v-card-text>
    <v-card-text v-if="total > limit" class="text-caption">
      Showing {{ items.length }} of {{ total }}
    </v-card-text>
  </v-card>
</template>

<style scoped>
.refresh-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
}

.refresh-card {
  background-color: rgba(var(--v-theme-on-surface), 0.03);
  border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
  border-radius: 10px;
  padding: 12px 14px;
}

.refresh-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.refresh-card__when {
  font-size: 0.8125rem;
  color: rgba(var(--v-theme-on-surface), 0.65);
}

.refresh-card__stats {
  display: flex;
  gap: 18px;
  margin-top: 10px;
}

.refresh-card__stat {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.refresh-card__stat-label {
  font-size: 0.6875rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: rgba(var(--v-theme-on-surface), 0.55);
}

.refresh-card__stat-value {
  font-size: 0.9375rem;
  font-weight: 500;
}

.refresh-card__bottom {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  flex-wrap: wrap;
}

.refresh-card__error {
  font-size: 0.8125rem;
  color: rgb(var(--v-theme-error));
}
</style>
