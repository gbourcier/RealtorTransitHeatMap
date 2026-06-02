<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import cronstrue from 'cronstrue'
import type { JobType, Schedule } from '../api/schedules'
import { BUILDING_TYPES, BED_BATH_OPTIONS } from '../constants/realtor'

const props = defineProps<{
  modelValue: boolean
  schedule: Schedule | null
  jobType?: JobType
}>()

interface SubmitPayload {
  name: string
  cronExpr: string
  enabled: boolean
  buildingTypeId?: number | null
  bedRange?: string | null
  bathRange?: string | null
  priceMin?: number | null
  priceMax?: number | null
  polygonWkt?: string | null
}

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', payload: SubmitPayload): void
}>()

const name = ref('')
const cronExpr = ref('0 */6 * * *')
const enabled = ref(true)
const submitting = ref(false)

const buildingTypeId = ref<number>(1)
const bedRange = ref('')
const bathRange = ref('')
const priceMin = ref('')
const priceMax = ref('')
const polygonWkt = ref('')

const effectiveJobType = computed(() => props.schedule?.jobType ?? props.jobType)
const isScrape = computed(() => effectiveJobType.value === 'scrape_realtor')

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      const s = props.schedule
      name.value = s?.name ?? ''
      cronExpr.value = s?.cronExpr ?? '0 */6 * * *'
      enabled.value = s?.enabled ?? true
      submitting.value = false

      buildingTypeId.value = s?.buildingTypeId ?? 1
      bedRange.value = s?.bedRange ?? ''
      bathRange.value = s?.bathRange ?? ''
      priceMin.value = s?.priceMin != null ? String(s.priceMin) : ''
      priceMax.value = s?.priceMax != null ? String(s.priceMax) : ''
      polygonWkt.value = s?.polygonWkt ?? ''
    }
  },
)

const cronHumanized = computed(() => {
  try {
    return cronstrue.toString(cronExpr.value, { use24HourTimeFormat: true })
  } catch {
    return ''
  }
})

const cronValid = computed(() => cronHumanized.value !== '')

const priceValid = computed(() => {
  if (priceMin.value === '' || priceMax.value === '') return true
  return Number(priceMin.value) <= Number(priceMax.value)
})
const polygonValid = computed(() => !isScrape.value || polygonWkt.value.trim().length > 0)

const formValid = computed(
  () =>
    name.value.trim().length > 0 &&
    cronValid.value &&
    polygonValid.value &&
    priceValid.value,
)

function close() {
  emit('update:modelValue', false)
}

function toPrice(v: string): number | null {
  return v === '' ? null : Number(v)
}

async function onSubmit() {
  if (!formValid.value) return
  submitting.value = true
  const payload: SubmitPayload = {
    name: name.value.trim(),
    cronExpr: cronExpr.value.trim(),
    enabled: enabled.value,
  }
  if (isScrape.value) {
    payload.buildingTypeId = buildingTypeId.value
    payload.bedRange = bedRange.value
    payload.bathRange = bathRange.value
    payload.priceMin = toPrice(priceMin.value)
    payload.priceMax = toPrice(priceMax.value)
    payload.polygonWkt = polygonWkt.value.trim()
  }
  emit('submit', payload)
}
</script>

<template>
  <v-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)" max-width="600">
    <v-card>
      <v-card-title class="d-flex align-center">
        <span>{{ schedule ? 'Edit schedule' : 'New schedule' }}</span>
        <v-spacer />
        <v-btn icon="mdi-close" variant="text" size="small" @click="close" />
      </v-card-title>
      <v-card-text>
        <v-text-field v-model="name" label="Name" variant="outlined" density="comfortable" />
        <v-text-field
          v-model="cronExpr"
          label="Cron expression (UTC)"
          hint="5 fields: min hour dom mon dow"
          persistent-hint
          variant="outlined"
          density="comfortable"
          :error="!cronValid && cronExpr.length > 0"
          :messages="cronValid ? cronHumanized : 'invalid cron'"
        />

        <template v-if="isScrape">
          <v-divider class="my-4" />
          <div class="text-subtitle-2 mb-2">Search filters</div>
          <v-select
            v-model="buildingTypeId"
            :items="BUILDING_TYPES"
            item-title="label"
            item-value="id"
            label="Property type"
            variant="outlined"
            density="comfortable"
          />
          <div class="d-flex ga-3">
            <v-select
              v-model="bedRange"
              :items="BED_BATH_OPTIONS"
              item-title="label"
              item-value="value"
              label="Beds"
              variant="outlined"
              density="comfortable"
            />
            <v-select
              v-model="bathRange"
              :items="BED_BATH_OPTIONS"
              item-title="label"
              item-value="value"
              label="Baths"
              variant="outlined"
              density="comfortable"
            />
          </div>
          <div class="d-flex ga-3">
            <v-text-field
              v-model="priceMin"
              type="number"
              label="Min price (CAD)"
              variant="outlined"
              density="comfortable"
              :error="!priceValid"
            />
            <v-text-field
              v-model="priceMax"
              type="number"
              label="Max price (CAD)"
              variant="outlined"
              density="comfortable"
              :error="!priceValid"
              :messages="priceValid ? '' : 'min must be ≤ max'"
            />
          </div>
          <v-textarea
            v-model="polygonWkt"
            label="Search area (WKT polygon)"
            hint="MULTIPOLYGON (((lon lat, …)))"
            persistent-hint
            variant="outlined"
            density="comfortable"
            rows="4"
            auto-grow
            :error="!polygonValid"
            :messages="polygonValid ? '' : 'required'"
          />
        </template>

        <v-switch v-model="enabled" label="Enabled" color="primary" hide-details class="mt-2" />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="close">Cancel</v-btn>
        <v-btn color="primary" variant="flat" :disabled="!formValid || submitting" :loading="submitting" @click="onSubmit">
          {{ schedule ? 'Save' : 'Create' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
