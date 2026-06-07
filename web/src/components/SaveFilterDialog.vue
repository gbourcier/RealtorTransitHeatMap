<script setup lang="ts">
import { computed, nextTick, ref, watch } from "vue";
import type { ListingFiltersState } from "../composables/useListingFilters";
import type { SavedViewsState } from "../composables/useSavedViews";
import { formatCompactPrice } from "../utils/listingFormat";

interface Props {
    modelValue: boolean;
    filters: ListingFiltersState;
    views: SavedViewsState;
}

const props = defineProps<Props>();
const emit = defineEmits<{
    "update:modelValue": [value: boolean];
}>();

const open = computed({
    get: () => props.modelValue,
    set: (v: boolean) => emit("update:modelValue", v),
});

const name = ref("");
const saving = ref(false);
const saveError = ref<string | null>(null);
const nameInput = ref<HTMLInputElement | null>(null);

const suggested = computed(() => {
    const parts: string[] = [];
    if (props.filters.maxCommuteSec.value != null) {
        parts.push(`≤${Math.round(props.filters.maxCommuteSec.value / 60)}m`);
    }
    if (props.filters.maxPrice.value != null) {
        parts.push(`under ${formatCompactPrice(props.filters.maxPrice.value)}`);
    }
    if (props.filters.minBedrooms.value != null) {
        parts.push(`${props.filters.minBedrooms.value}+ bd`);
    }
    return parts.join(" · ") || "My filter";
});

watch(
    open,
    (isOpen) => {
        if (!isOpen) return;
        name.value = suggested.value;
        saveError.value = null;
        nextTick(() => nameInput.value?.focus());
    },
    { immediate: true },
);

async function commitSave(): Promise<void> {
    const trimmed = name.value.trim();
    if (!trimmed || saving.value) return;
    saving.value = true;
    saveError.value = null;
    try {
        await props.views.saveAsNew(trimmed);
        open.value = false;
    } catch (e: any) {
        saveError.value =
            e?.response?.status === 409
                ? "A saved filter with that name already exists."
                : "Failed to save filter.";
    } finally {
        saving.value = false;
    }
}
</script>

<template>
    <Teleport to="body">
        <div v-if="open" class="save-filter" @click="open = false">
            <div class="save-filter__card" role="dialog" aria-modal="true" aria-label="Save filter" @click.stop>
                <div class="save-filter__title">Save these filters as a view</div>
                <div class="save-filter__sub">
                    It'll appear in the header view menu for one-tap recall.
                </div>
                <input
                    ref="nameInput"
                    v-model="name"
                    class="save-filter__input"
                    placeholder="e.g. Downtown commute"
                    maxlength="60"
                    @keydown.enter="commitSave"
                    @keydown.esc="open = false"
                />
                <p v-if="saveError" class="save-filter__error">{{ saveError }}</p>
                <div class="save-filter__actions">
                    <button class="save-filter__cancel" @click="open = false">Cancel</button>
                    <button
                        class="save-filter__ok"
                        :disabled="!name.trim() || saving"
                        @click="commitSave"
                    >Save view</button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<style scoped>
.save-filter {
    position: fixed;
    inset: 0;
    z-index: 2000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 22px;
    background: rgba(var(--v-theme-map-bg), 0.62);
    backdrop-filter: blur(3px);
}

.save-filter__card {
    width: min(320px, 100%);
    padding: 22px;
    border-radius: 16px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    box-shadow: 0 18px 50px rgba(var(--v-theme-shadow), 0.55);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.save-filter__title {
    font-size: 1.0625rem;
    font-weight: 700;
}

.save-filter__sub {
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin: 5px 0 16px;
}

.save-filter__input {
    width: 100%;
    height: 44px;
    padding: 0 13px;
    border-radius: 10px;
    background: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.18);
    color: rgba(var(--v-theme-on-surface), 0.95);
    font-size: 0.875rem;
    outline: none;
    transition: border-color 120ms ease, background-color 120ms ease;
}

.save-filter__input::placeholder {
    color: rgba(var(--v-theme-on-surface), 0.4);
}

.save-filter__input:focus {
    border-color: rgb(var(--v-theme-primary));
    background: rgba(var(--v-theme-on-surface), 0.09);
}

.save-filter__error {
    margin: 8px 0 0;
    font-size: 0.75rem;
    color: rgb(var(--v-theme-error));
}

.save-filter__actions {
    display: flex;
    gap: 10px;
    margin-top: 18px;
}

.save-filter__cancel {
    flex: 1 1 0;
    height: 42px;
    border-radius: 999px;
    cursor: pointer;
    background: transparent;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.2);
    color: rgba(var(--v-theme-on-surface), 0.9);
    font-size: 0.875rem;
    font-weight: 500;
}

.save-filter__cancel:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
}

.save-filter__ok {
    flex: 1 1 0;
    height: 42px;
    border-radius: 999px;
    cursor: pointer;
    background: rgb(var(--v-theme-primary));
    border: 0;
    color: rgb(var(--v-theme-on-primary));
    font-size: 0.875rem;
    font-weight: 700;
    transition: filter 120ms ease, opacity 120ms ease;
}

.save-filter__ok:hover:not(:disabled) {
    filter: brightness(1.06);
}

.save-filter__ok:disabled {
    opacity: 0.4;
    cursor: default;
}

.save-filter__cancel:focus-visible,
.save-filter__ok:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}
</style>
