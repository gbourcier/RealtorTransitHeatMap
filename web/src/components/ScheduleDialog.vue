<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import cronstrue from 'cronstrue'
import type { Schedule } from '../api/schedules'

const props = defineProps<{
  modelValue: boolean
  schedule: Schedule | null
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void
  (e: 'submit', payload: { name: string; cronExpr: string; enabled: boolean }): void
}>()

const name = ref('')
const cronExpr = ref('0 */6 * * *')
const enabled = ref(true)
const submitting = ref(false)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      name.value = props.schedule?.name ?? ''
      cronExpr.value = props.schedule?.cronExpr ?? '0 */6 * * *'
      enabled.value = props.schedule?.enabled ?? true
      submitting.value = false
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
const formValid = computed(() => name.value.trim().length > 0 && cronValid.value)

function close() {
  emit('update:modelValue', false)
}

async function onSubmit() {
  if (!formValid.value) return
  submitting.value = true
  emit('submit', {
    name: name.value.trim(),
    cronExpr: cronExpr.value.trim(),
    enabled: enabled.value,
  })
}
</script>

<template>
  <v-dialog :model-value="modelValue" @update:model-value="emit('update:modelValue', $event)" max-width="500">
    <v-card>
      <v-card-title>{{ schedule ? 'Edit schedule' : 'New schedule' }}</v-card-title>
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
        <v-switch v-model="enabled" label="Enabled" color="primary" hide-details class="mt-2" />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" @click="close">Cancel</v-btn>
        <v-btn color="primary" :disabled="!formValid || submitting" :loading="submitting" @click="onSubmit">
          {{ schedule ? 'Save' : 'Create' }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
