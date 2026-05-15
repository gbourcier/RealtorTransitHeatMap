<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { listRuns, startScrape, type ScrapeRun, type ScrapeStatus } from '../api/scrapes'

const items = ref<ScrapeRun[]>([])
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
    const res = await listRuns({ limit: limit.value, offset: offset.value })
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
    await startScrape()
    await refresh()
  } catch (e: any) {
    if (e?.response?.status === 409) {
      error.value = 'A scrape is already in progress.'
    } else {
      error.value = e?.response?.data?.error ?? e?.message ?? 'trigger failed'
    }
  } finally {
    triggering.value = false
  }
}

function statusColor(s: ScrapeStatus): string {
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

function durationLabel(r: ScrapeRun): string {
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
      <span>Scrape Jobs</span>
      <v-spacer />
      <v-btn variant="text" icon="mdi-refresh" :loading="loading" @click="refresh" />
      <v-btn
        color="primary"
        prepend-icon="mdi-play"
        :loading="triggering"
        :disabled="hasRunning"
        @click="onTrigger"
      >
        Run scrape now
      </v-btn>
    </v-card-title>
    <v-divider />
    <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{ error }}</v-alert>
    <v-card-text v-if="loading && items.length === 0" class="text-center py-8">
      <v-progress-circular indeterminate />
    </v-card-text>
    <v-table v-else-if="items.length > 0" density="comfortable">
      <thead>
        <tr>
          <th>Status</th>
          <th>Started</th>
          <th>Duration</th>
          <th class="text-right">Total</th>
          <th class="text-right">New</th>
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
          <td class="text-right">{{ r.totalCount ?? '—' }}</td>
          <td class="text-right">{{ r.newCount ?? '—' }}</td>
          <td>
            <v-chip v-if="r.scheduleId" size="x-small" variant="outlined">scheduled</v-chip>
            <v-chip v-else size="x-small" variant="outlined" color="grey">manual</v-chip>
          </td>
          <td class="text-caption">
            <span v-if="r.errorKind">{{ r.errorKind }}: {{ r.errorMessage }}</span>
          </td>
        </tr>
      </tbody>
    </v-table>
    <v-card-text v-else class="text-medium-emphasis text-center py-8">
      No runs yet. Click "Run scrape now" to start one.
    </v-card-text>
    <v-card-text v-if="total > limit" class="text-caption">
      Showing {{ items.length }} of {{ total }}
    </v-card-text>
  </v-card>
</template>
