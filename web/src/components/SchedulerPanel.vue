<script setup lang="ts">
import { onMounted, ref } from 'vue'
import cronstrue from 'cronstrue'
import {
  listSchedules,
  createSchedule,
  updateSchedule,
  deleteSchedule,
  type JobType,
  type Schedule,
} from '../api/schedules'
import ScheduleDialog from './ScheduleDialog.vue'

const props = defineProps<{
  jobType?: JobType
}>()

const items = ref<Schedule[]>([])
const total = ref(0)
const limit = ref(50)
const offset = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)

const dialogOpen = ref(false)
const editing = ref<Schedule | null>(null)

async function refresh() {
  loading.value = true
  error.value = null
  try {
    const res = await listSchedules({
      limit: limit.value,
      offset: offset.value,
      ...(props.jobType ? { jobType: props.jobType } : {}),
    })
    items.value = res.items
    total.value = res.total
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? e?.message ?? 'failed to load schedules'
  } finally {
    loading.value = false
  }
}

onMounted(refresh)

function humanize(expr: string): string {
  try {
    return cronstrue.toString(expr, { use24HourTimeFormat: true })
  } catch {
    return '—'
  }
}

function openCreate() {
  editing.value = null
  dialogOpen.value = true
}

function openEdit(s: Schedule) {
  editing.value = s
  dialogOpen.value = true
}

async function onSubmit(payload: { name: string; cronExpr: string; enabled: boolean }) {
  try {
    if (editing.value) {
      await updateSchedule(editing.value.id, payload)
    } else {
      await createSchedule({ ...payload, ...(props.jobType ? { jobType: props.jobType } : {}) })
    }
    dialogOpen.value = false
    await refresh()
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? e?.message ?? 'save failed'
  }
}

async function toggleEnabled(s: Schedule) {
  try {
    await updateSchedule(s.id, { enabled: !s.enabled })
    await refresh()
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? e?.message ?? 'toggle failed'
  }
}

async function onDelete(s: Schedule) {
  if (!confirm(`Delete schedule "${s.name}"?`)) return
  try {
    await deleteSchedule(s.id)
    await refresh()
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? e?.message ?? 'delete failed'
  }
}
</script>

<template>
  <v-card>
    <v-card-title class="d-flex align-center">
      <span>{{ 'Schedules' }}</span>
      <v-spacer />
      <v-btn color="primary" prepend-icon="mdi-plus" @click="openCreate">New</v-btn>
    </v-card-title>
    <v-divider />
    <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{ error }}</v-alert>
    <v-card-text v-if="loading && items.length === 0" class="text-center py-8">
      <v-progress-circular indeterminate />
    </v-card-text>
    <v-list v-else-if="items.length > 0" lines="two">
      <template v-for="(s, i) in items" :key="s.id">
        <v-list-item>
          <template #prepend>
            <v-switch :model-value="s.enabled" hide-details density="compact" color="primary" class="mr-4"
              style="margin-top: 0; margin-bottom: 0;" @update:model-value="toggleEnabled(s)" />
          </template>
          <v-list-item-title>{{ s.name }}</v-list-item-title>
          <v-list-item-subtitle>
            <code>{{ s.cronExpr }}</code> — {{ humanize(s.cronExpr) }}
          </v-list-item-subtitle>
          <template #append>
            <v-btn icon="mdi-pencil" variant="text" size="small" @click="openEdit(s)" />
            <v-btn icon="mdi-delete" variant="text" size="small" color="error" @click="onDelete(s)" />
          </template>
        </v-list-item>
        <v-divider v-if="i < items.length - 1" />
      </template>
    </v-list>
    <v-card-text v-else class="text-medium-emphasis text-center py-8">
      No schedules yet. Click "New" to create one.
    </v-card-text>
    <v-card-text v-if="total > limit" class="text-caption">
      Showing {{ items.length }} of {{ total }}
    </v-card-text>

    <ScheduleDialog v-model="dialogOpen" :schedule="editing" @submit="onSubmit" />
  </v-card>
</template>
