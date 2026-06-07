<script setup lang="ts">
import { ref, computed, watch, nextTick } from "vue";
import type { ListingFiltersState } from "../composables/useListingFilters";
import type { SavedViewsState } from "../composables/useSavedViews";
import { formatCompactPrice } from "../utils/listingFormat";
import HeatSlider from "./HeatSlider.vue";

interface Props {
    state: ListingFiltersState;
    views: SavedViewsState;
    total: number;
    saveOpen: boolean;
}

const props = defineProps<Props>();
const emit = defineEmits<{
    close: [];
    error: [message: string];
    "update:saveOpen": [value: boolean];
}>();

const PRICE_MIN = 300_000;
const PRICE_MAX = 2_000_000;
const PRICE_STEP = 25_000;
const PRICE_NO_MAX = PRICE_MAX + PRICE_STEP;
const priceTicks = ["$300k", "$1M", "Any"];

const COMMUTE_MAX = 60;
const COMMUTE_STEP = 5;
const COMMUTE_NO_MAX = COMMUTE_MAX + COMMUTE_STEP;
const commuteTicks = ["0m", "30m", "Any"];

const bedOptions = [
    { value: null, label: "Any" },
    { value: 1, label: "1+" },
    { value: 2, label: "2+" },
    { value: 3, label: "3+" },
    { value: 4, label: "4+" },
];
const bathOptions = [
    { value: null, label: "Any" },
    { value: 1, label: "1+" },
    { value: 2, label: "2+" },
    { value: 3, label: "3+" },
];
const recencyOptions = [
    { value: null, label: "Any time" },
    { value: 1, label: "Today" },
    { value: 7, label: "This week" },
];

const commuteMin = computed(() =>
    props.state.maxCommuteSec.value == null
        ? null
        : Math.round(props.state.maxCommuteSec.value / 60),
);

const commuteSlider = computed({
    get: () => commuteMin.value ?? COMMUTE_NO_MAX,
    set: (v: number) => {
        props.state.maxCommuteSec.value = v >= COMMUTE_NO_MAX ? null : v * 60;
    },
});

const priceSlider = computed({
    get: () => props.state.maxPrice.value ?? PRICE_NO_MAX,
    set: (v: number) => {
        props.state.maxPrice.value = v >= PRICE_NO_MAX ? null : v;
    },
});

const priceLabel = computed(() =>
    props.state.maxPrice.value == null
        ? "Any"
        : `≤ ${formatCompactPrice(props.state.maxPrice.value)}`,
);

const saveOpenLocal = computed({
    get: () => props.saveOpen,
    set: (v: boolean) => emit("update:saveOpen", v),
});

const name = ref("");
const saving = ref(false);
const saveError = ref<string | null>(null);
const nameInput = ref<HTMLInputElement | null>(null);

const suggested = computed(() => {
    const parts: string[] = [];
    if (props.state.maxCommuteSec.value != null) {
        parts.push(`≤${Math.round(props.state.maxCommuteSec.value / 60)}m`);
    }
    if (props.state.maxPrice.value != null) {
        parts.push(`under ${formatCompactPrice(props.state.maxPrice.value)}`);
    }
    if (props.state.minBedrooms.value != null) {
        parts.push(`${props.state.minBedrooms.value}+ bd`);
    }
    return parts.join(" · ") || "My filter";
});

watch(
    saveOpenLocal,
    (open) => {
        if (!open) return;
        name.value = suggested.value;
        saveError.value = null;
        nextTick(() => nameInput.value?.focus());
    },
    { immediate: true },
);

function openSave(): void {
    saveOpenLocal.value = true;
}

async function commitSave(): Promise<void> {
    const trimmed = name.value.trim();
    if (!trimmed || saving.value) return;
    saving.value = true;
    saveError.value = null;
    try {
        await props.views.saveAsNew(trimmed);
        saveOpenLocal.value = false;
    } catch (e: any) {
        saveError.value =
            e?.response?.status === 409
                ? "A saved filter with that name already exists."
                : "Failed to save filter.";
    } finally {
        saving.value = false;
    }
}

async function onUpdate(): Promise<void> {
    try {
        await props.views.updateActive();
    } catch {
        emit("error", "Failed to update saved view.");
    }
}
</script>

<template>
    <div class="filter-drawer" role="dialog" aria-label="Filters">
        <header class="fp__head">
            <span class="fp__title">Filters</span>
            <span v-if="state.activeFilterCount.value > 0" class="fp__badge">
                {{ state.activeFilterCount.value }} active
            </span>
            <button class="fp__x" aria-label="Close filters" @click="emit('close')">
                <v-icon size="20">mdi-close</v-icon>
            </button>
        </header>

        <div v-if="views.active.value" class="fp__applied">
            <v-icon size="15" class="fp__applied__check">mdi-check</v-icon>
            <span class="fp__applied__text">
                <span class="fp__applied__name">{{ views.active.value.name }}</span>
                <span
                    class="fp__applied__state"
                    :class="{ 'fp__applied__state--dirty': views.dirty.value }"
                >
                    {{ views.dirty.value ? "· edited" : "· saved view" }}
                </span>
            </span>
            <span v-if="views.dirty.value" class="fp__applied__actions">
                <button
                    class="fp__applied__btn fp__applied__btn--ghost"
                    @click="openSave"
                >Save as new</button>
                <button
                    class="fp__applied__btn fp__applied__btn--primary"
                    @click="onUpdate"
                >Update</button>
            </span>
        </div>

        <div class="fp__body">
            <div class="hero">
                <div class="hero__top">
                    <span class="hero__kicker">Commute to downtown</span>
                    <span class="hero__read">
                        <template v-if="commuteMin == null">Any</template>
                        <template v-else>{{ commuteMin }}<small>min</small></template>
                    </span>
                </div>
                <HeatSlider
                    v-model="commuteSlider"
                    :min="0"
                    :max="COMMUTE_NO_MAX"
                    :step="COMMUTE_STEP"
                    :ticks="commuteTicks"
                    heat
                    aria-label="Maximum commute to downtown"
                />
                <div class="hero__counts">
                    <span class="hero__total">
                        {{ total.toLocaleString() }}<small>in reach</small>
                    </span>
                </div>
            </div>

            <section class="fp__sec">
                <div class="fp__sec-head">
                    <span class="fp__label">Max price</span>
                    <span class="fp__val" :class="{ 'fp__val--set': state.maxPrice.value != null }">
                        {{ priceLabel }}
                    </span>
                </div>
                <HeatSlider
                    v-model="priceSlider"
                    :min="PRICE_MIN"
                    :max="PRICE_NO_MAX"
                    :step="PRICE_STEP"
                    :ticks="priceTicks"
                    aria-label="Maximum price"
                />
            </section>

            <section class="fp__sec">
                <div class="fp__sec-head"><span class="fp__label">Bedrooms</span></div>
                <div class="chips" role="radiogroup" aria-label="Minimum bedrooms">
                    <button
                        v-for="opt in bedOptions"
                        :key="String(opt.value)"
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': state.minBedrooms.value === opt.value }"
                        role="radio"
                        :aria-checked="state.minBedrooms.value === opt.value"
                        @click="state.minBedrooms.value = opt.value"
                    >{{ opt.label }}</button>
                </div>
                <div class="fp__sec-head fp__sec-head--inset">
                    <span class="fp__label">Bathrooms</span>
                </div>
                <div class="chips" role="radiogroup" aria-label="Minimum bathrooms">
                    <button
                        v-for="opt in bathOptions"
                        :key="String(opt.value)"
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': state.minBathrooms.value === opt.value }"
                        role="radio"
                        :aria-checked="state.minBathrooms.value === opt.value"
                        @click="state.minBathrooms.value = opt.value"
                    >{{ opt.label }}</button>
                </div>
            </section>

            <section class="fp__sec">
                <div class="fp__sec-head"><span class="fp__label">Freshness</span></div>
                <div class="chips">
                    <button
                        v-for="opt in recencyOptions"
                        :key="String(opt.value)"
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': state.newWithinDays.value === opt.value }"
                        @click="state.newWithinDays.value = opt.value"
                    >{{ opt.label }}</button>
                    <span class="chips__sep" />
                    <button
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': state.includeExpired.value }"
                        :aria-pressed="state.includeExpired.value"
                        @click="state.includeExpired.value = !state.includeExpired.value"
                    >
                        <v-icon v-if="state.includeExpired.value" size="13">mdi-check</v-icon>
                        Include expired
                    </button>
                </div>
            </section>
        </div>

        <footer class="fp__foot">
            <button class="fp__reset" @click="views.resetFilters()">Reset</button>
            <button class="fp__save" @click="openSave">
                <v-icon size="16">mdi-bookmark-outline</v-icon>
                Save
            </button>
            <button class="fp__apply" @click="emit('close')">
                Show <b>{{ total.toLocaleString() }}</b> listing{{ total === 1 ? "" : "s" }}
            </button>
        </footer>

        <div v-if="saveOpenLocal" class="fp__modal" @click="saveOpenLocal = false">
            <div class="fp__modal-card" @click.stop>
                <div class="fp__modal-title">Save these filters as a view</div>
                <div class="fp__modal-sub">
                    It'll appear in the header view menu for one-tap recall.
                </div>
                <input
                    ref="nameInput"
                    v-model="name"
                    class="fp__modal-input"
                    placeholder="e.g. Downtown commute"
                    maxlength="60"
                    @keydown.enter="commitSave"
                />
                <p v-if="saveError" class="fp__modal-error">{{ saveError }}</p>
                <div class="fp__modal-actions">
                    <button class="fp__modal-cancel" @click="saveOpenLocal = false">Cancel</button>
                    <button
                        class="fp__modal-ok"
                        :disabled="!name.trim() || saving"
                        @click="commitSave"
                    >Save view</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.filter-drawer {
    position: fixed;
    top: 62px;
    right: 14px;
    max-height: calc(100dvh - 76px);
    width: min(404px, calc(100vw - 28px));
    z-index: 1005;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: 18px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.07);
    box-shadow: 0 26px 72px rgba(var(--v-theme-shadow), 0.55);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.fp__head {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 18px 20px 12px;
}

.fp__title {
    font-size: 1.25rem;
    font-weight: 700;
    letter-spacing: -0.2px;
}

.fp__badge {
    display: inline-flex;
    align-items: center;
    height: 22px;
    padding: 0 9px;
    border-radius: 999px;
    font-size: 0.75rem;
    font-weight: 600;
    white-space: nowrap;
    color: rgb(var(--v-theme-primary));
    border: 1px solid rgba(var(--v-theme-primary), 0.5);
    background: rgba(var(--v-theme-primary), 0.07);
}

.fp__x {
    margin-left: auto;
    width: 32px;
    height: 32px;
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 0;
    border-radius: 999px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.55);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.fp__x:hover {
    background: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.fp__x:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: -2px;
}

.fp__applied {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 2px 14px 0;
    padding: 10px 14px;
    border-radius: 13px;
    background: rgba(var(--v-theme-primary), 0.09);
    border: 1px solid rgba(var(--v-theme-primary), 0.32);
}

.fp__applied__check {
    flex: 0 0 auto;
    color: rgb(var(--v-theme-primary));
}

.fp__applied__text {
    min-width: 0;
    display: inline-flex;
    align-items: baseline;
    gap: 5px;
    overflow: hidden;
}

.fp__applied__name {
    font-size: 0.84375rem;
    font-weight: 700;
    color: rgb(var(--v-theme-primary));
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.fp__applied__state {
    flex: 0 0 auto;
    font-size: 0.75rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.5);
}

.fp__applied__state--dirty {
    color: rgb(var(--v-theme-warning));
}

.fp__applied__actions {
    margin-left: auto;
    flex: 0 0 auto;
    display: flex;
    gap: 8px;
}

.fp__applied__btn {
    height: 30px;
    padding: 0 12px;
    border-radius: 999px;
    cursor: pointer;
    font-size: 0.78125rem;
    font-weight: 600;
    white-space: nowrap;
    transition: filter 120ms ease, background-color 120ms ease;
}

.fp__applied__btn--primary {
    background: rgb(var(--v-theme-primary));
    border: 0;
    color: rgb(var(--v-theme-on-primary));
}

.fp__applied__btn--primary:hover {
    filter: brightness(1.06);
}

.fp__applied__btn--ghost {
    background: transparent;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
    color: rgba(var(--v-theme-on-surface), 0.85);
}

.fp__applied__btn--ghost:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
}

.fp__body {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    overflow-x: hidden;
}

.fp__body::-webkit-scrollbar {
    width: 10px;
}

.fp__body::-webkit-scrollbar-thumb {
    background: rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 999px;
    border: 3px solid rgb(var(--v-theme-surface));
}

.hero {
    margin: 6px 14px 0;
    padding: 18px 16px;
    border-radius: 16px;
    background: rgb(var(--v-theme-map-bg));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.07);
}

.hero__top {
    display: flex;
    align-items: flex-end;
    justify-content: space-between;
    margin-bottom: 16px;
}

.hero__kicker {
    font-size: 0.6875rem;
    font-weight: 700;
    letter-spacing: 1.3px;
    text-transform: uppercase;
    color: rgb(var(--v-theme-primary));
}

.hero__read {
    font-size: 2.75rem;
    font-weight: 900;
    line-height: 0.9;
    letter-spacing: -1.5px;
    color: rgba(var(--v-theme-on-surface), 0.96);
    font-variant-numeric: tabular-nums;
}

.hero__read small {
    font-size: 1rem;
    font-weight: 500;
    margin-left: 5px;
    color: rgba(var(--v-theme-on-surface), 0.5);
    letter-spacing: 0;
}

.hero__counts {
    display: flex;
    margin-top: 16px;
}

.hero__total {
    margin-left: auto;
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    font-size: 1.1875rem;
    font-weight: 700;
    color: rgb(var(--v-theme-primary));
    font-variant-numeric: tabular-nums;
}

.hero__total small {
    font-size: 0.6875rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.5);
}

.fp__sec {
    padding: 16px 20px;
}

.fp__sec + .fp__sec {
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.fp__sec-head {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin-bottom: 12px;
}

.fp__sec-head--inset {
    margin-top: 16px;
}

.fp__label {
    font-size: 0.875rem;
    font-weight: 600;
}

.fp__val {
    margin-left: auto;
    font-size: 0.8125rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.62);
    font-variant-numeric: tabular-nums;
}

.fp__val--set {
    color: rgb(var(--v-theme-primary));
}

.chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
}

.chips__sep {
    flex-basis: 100%;
    height: 4px;
}

.chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    height: 34px;
    padding: 0 13px;
    border-radius: 999px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.18);
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.82);
    white-space: nowrap;
    font-size: 0.8125rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.chip:hover {
    background: rgba(var(--v-theme-on-surface), 0.05);
    border-color: rgba(var(--v-theme-on-surface), 0.3);
}

.chip:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.chip--on,
.chip--on:hover {
    background: rgba(var(--v-theme-primary), 0.16);
    border-color: rgba(var(--v-theme-primary), 0.55);
    color: rgb(var(--v-theme-primary));
    font-weight: 700;
}

.fp__foot {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 20px 16px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.07);
}

.fp__reset {
    flex: 0 0 auto;
    height: 44px;
    padding: 0 18px;
    border-radius: 999px;
    background: transparent;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.2);
    color: rgba(var(--v-theme-on-surface), 0.9);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.fp__reset:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.fp__save {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 44px;
    padding: 0 15px;
    border-radius: 999px;
    background: transparent;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.2);
    color: rgba(var(--v-theme-on-surface), 0.9);
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.fp__save:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.fp__reset:focus-visible,
.fp__save:focus-visible,
.fp__apply:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.fp__apply {
    flex: 1 1 auto;
    height: 44px;
    padding: 0 18px;
    border-radius: 999px;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    border: 0;
    font-size: 0.9375rem;
    font-weight: 700;
    letter-spacing: -0.1px;
    cursor: pointer;
    transition: filter 120ms ease;
}

.fp__apply:hover {
    filter: brightness(1.06);
}

.fp__modal {
    position: absolute;
    inset: 0;
    z-index: 20;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 22px;
    background: rgba(var(--v-theme-map-bg), 0.62);
    backdrop-filter: blur(3px);
}

.fp__modal-card {
    width: 100%;
    max-width: 320px;
    padding: 22px;
    border-radius: 16px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    box-shadow: 0 18px 50px rgba(var(--v-theme-shadow), 0.55);
}

.fp__modal-title {
    font-size: 1.0625rem;
    font-weight: 700;
}

.fp__modal-sub {
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin: 5px 0 16px;
}

.fp__modal-input {
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

.fp__modal-input::placeholder {
    color: rgba(var(--v-theme-on-surface), 0.4);
}

.fp__modal-input:focus {
    border-color: rgb(var(--v-theme-primary));
    background: rgba(var(--v-theme-on-surface), 0.09);
}

.fp__modal-error {
    margin: 8px 0 0;
    font-size: 0.75rem;
    color: rgb(var(--v-theme-error));
}

.fp__modal-actions {
    display: flex;
    gap: 10px;
    margin-top: 18px;
}

.fp__modal-cancel {
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

.fp__modal-cancel:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
}

.fp__modal-ok {
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

.fp__modal-ok:hover:not(:disabled) {
    filter: brightness(1.06);
}

.fp__modal-ok:disabled {
    opacity: 0.4;
    cursor: default;
}

@media (max-width: 600px) {
    .filter-drawer {
        top: 58px;
        left: 12px;
        right: 12px;
        max-height: calc(100dvh - 70px);
        width: auto;
    }
}
</style>
