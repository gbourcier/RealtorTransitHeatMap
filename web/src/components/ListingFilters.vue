<script setup lang="ts">
import { ref, computed } from "vue";
import type { ListingFiltersState } from "../composables/useListingFilters";
import { formatCompactPrice } from "../utils/listingFormat";

interface Props {
    state: ListingFiltersState;
    total: number;
}

const props = defineProps<Props>();

const bedroomOptions = [1, 2, 3, 4];
const bathroomOptions = [1, 2, 3];
const recencyOptions: { label: string; days: number | null }[] = [
    { label: "All time", days: null },
    { label: "Today", days: 1 },
    { label: "This week", days: 7 },
];

const PRICE_MIN = 300_000;
const PRICE_MAX = 2_000_000;
const PRICE_STEP = 25_000;
const PRICE_NO_MAX = PRICE_MAX + PRICE_STEP;
const priceTicks: Record<number, string> = {
    [PRICE_MIN]: "300k",
    1_000_000: "1M",
    [PRICE_NO_MAX]: "Any",
};

const COMMUTE_MIN = 0;
const COMMUTE_MAX = 60;
const COMMUTE_STEP = 5;
const COMMUTE_NO_MAX = COMMUTE_MAX + COMMUTE_STEP;
const commuteTicks: Record<number, string> = {
    [COMMUTE_MIN]: "0m",
    30: "30m",
    [COMMUTE_NO_MAX]: "Any",
};

const SQFT_MIN = 0;
const SQFT_MAX = 2500;
const SQFT_STEP = 100;
const sqftTicks: Record<number, string> = {
    [SQFT_MIN]: "Any",
    1000: "1k",
    [SQFT_MAX]: "2.5k",
};

const menuOpen = ref(false);

const priceSlider = computed({
    get: () => props.state.maxPrice.value ?? PRICE_NO_MAX,
    set: (v: number) => {
        props.state.maxPrice.value = v >= PRICE_NO_MAX ? null : v;
    },
});

const commuteSliderMin = computed({
    get: () =>
        props.state.maxCommuteSec.value == null
            ? COMMUTE_NO_MAX
            : Math.round(props.state.maxCommuteSec.value / 60),
    set: (v: number) => {
        props.state.maxCommuteSec.value = v >= COMMUTE_NO_MAX ? null : v * 60;
    },
});

const sqftSlider = computed({
    get: () => props.state.minInteriorAreaSqft.value ?? SQFT_MIN,
    set: (v: number) => {
        props.state.minInteriorAreaSqft.value = v <= SQFT_MIN ? null : v;
    },
});
</script>

<template>
    <v-menu
        v-model="menuOpen"
        :close-on-content-click="false"
        location="bottom end"
        offset="18"
        transition="scale-transition"
    >
        <template #activator="{ props: activatorProps }">
            <button
                v-bind="activatorProps"
                type="button"
                class="filter-pill"
                :class="{
                    'filter-pill--active': state.activeFilterCount.value > 0,
                    'filter-pill--open': menuOpen,
                }"
            >
                <v-icon size="14" class="filter-pill__icon">mdi-tune-variant</v-icon>
                <span class="filter-pill__label">Filters</span>
                <template v-if="state.activeFilterCount.value > 0">
                    <span class="filter-pill__dot" aria-hidden="true">•</span>
                    <span class="filter-pill__count">{{ state.activeFilterCount.value }}</span>
                </template>
            </button>
        </template>
        <div class="filter-modal" role="dialog" aria-label="Filters">
            <header class="filter-modal__header">
                <div class="filter-modal__title-row">
                    <span class="filter-modal__title">Filters</span>
                    <span v-if="state.activeFilterCount.value > 0" class="filter-modal__active-badge">
                        {{ state.activeFilterCount.value }} active
                    </span>
                </div>
            </header>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Max price</span>
                    <span class="filter-modal__section-value">
                        {{ state.maxPrice.value == null ? "Any" : formatCompactPrice(state.maxPrice.value) }}
                    </span>
                    <button
                        v-if="state.maxPrice.value != null"
                        type="button"
                        class="filter-modal__section-clear"
                        @click="state.maxPrice.value = null"
                    >Clear</button>
                </div>
                <v-slider
                    v-model="priceSlider"
                    :min="PRICE_MIN"
                    :max="PRICE_NO_MAX"
                    :step="PRICE_STEP"
                    :ticks="priceTicks"
                    show-ticks="always"
                    tick-size="3"
                    color="secondary"
                    track-color="rgba(var(--v-theme-on-surface), 0.16)"
                    hide-details
                    density="compact"
                    class="filter-slider"
                />
            </section>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Max commute</span>
                    <span class="filter-modal__section-value">
                        {{ state.maxCommuteSec.value == null ? "Any" : `${Math.round(state.maxCommuteSec.value / 60)} min` }}
                    </span>
                    <button
                        v-if="state.maxCommuteSec.value != null"
                        type="button"
                        class="filter-modal__section-clear"
                        @click="state.maxCommuteSec.value = null"
                    >Clear</button>
                </div>
                <v-slider
                    v-model="commuteSliderMin"
                    :min="COMMUTE_MIN"
                    :max="COMMUTE_NO_MAX"
                    :step="COMMUTE_STEP"
                    :ticks="commuteTicks"
                    show-ticks="always"
                    tick-size="3"
                    color="secondary"
                    track-color="rgba(var(--v-theme-on-surface), 0.16)"
                    hide-details
                    density="compact"
                    class="filter-slider"
                />
            </section>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Min interior space</span>
                    <span class="filter-modal__section-value">
                        {{ state.minInteriorAreaSqft.value == null ? "Any" : `${state.minInteriorAreaSqft.value.toLocaleString()} sqft` }}
                    </span>
                    <button
                        v-if="state.minInteriorAreaSqft.value != null"
                        type="button"
                        class="filter-modal__section-clear"
                        @click="state.minInteriorAreaSqft.value = null"
                    >Clear</button>
                </div>
                <v-slider
                    v-model="sqftSlider"
                    :min="SQFT_MIN"
                    :max="SQFT_MAX"
                    :step="SQFT_STEP"
                    :ticks="sqftTicks"
                    show-ticks="always"
                    tick-size="3"
                    color="secondary"
                    track-color="rgba(var(--v-theme-on-surface), 0.16)"
                    hide-details
                    density="compact"
                    class="filter-slider"
                />
            </section>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Bedrooms</span>
                </div>
                <div class="filter-segmented" role="radiogroup" aria-label="Minimum bedrooms">
                    <button
                        v-for="b in bedroomOptions"
                        :key="b"
                        type="button"
                        class="filter-segmented__btn"
                        :class="{ 'filter-segmented__btn--active': state.minBedrooms.value === b }"
                        role="radio"
                        :aria-checked="state.minBedrooms.value === b"
                        @click="state.minBedrooms.value = b"
                    >
                        {{ b }}+
                    </button>
                </div>
            </section>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Bathrooms</span>
                </div>
                <div class="filter-segmented" role="radiogroup" aria-label="Minimum bathrooms">
                    <button
                        v-for="b in bathroomOptions"
                        :key="b"
                        type="button"
                        class="filter-segmented__btn"
                        :class="{ 'filter-segmented__btn--active': state.minBathrooms.value === b }"
                        role="radio"
                        :aria-checked="state.minBathrooms.value === b"
                        @click="state.minBathrooms.value = b"
                    >
                        {{ b }}+
                    </button>
                </div>
            </section>

            <section class="filter-modal__section">
                <div class="filter-modal__section-head">
                    <span class="filter-modal__section-title">Recency</span>
                </div>
                <div class="filter-segmented" role="radiogroup" aria-label="Listing recency">
                    <button
                        v-for="opt in recencyOptions"
                        :key="opt.label"
                        type="button"
                        class="filter-segmented__btn"
                        :class="{ 'filter-segmented__btn--active': state.newWithinDays.value === opt.days }"
                        role="radio"
                        :aria-checked="state.newWithinDays.value === opt.days"
                        @click="state.newWithinDays.value = opt.days"
                    >
                        {{ opt.label }}
                    </button>
                </div>
            </section>

            <footer class="filter-modal__footer">
                <button
                    type="button"
                    class="filter-modal__reset"
                    :disabled="state.activeFilterCount.value === 0"
                    @click="state.clearAll()"
                >Reset</button>
                <button
                    type="button"
                    class="filter-modal__apply"
                    @click="menuOpen = false"
                >
                    Show {{ total.toLocaleString() }} listing{{ total === 1 ? "" : "s" }}
                </button>
            </footer>
        </div>
    </v-menu>
</template>

<style scoped>
.filter-pill {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    height: 28px;
    padding: 0 10px;
    border-radius: 999px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.88);
    font-size: 0.75rem;
    font-weight: 500;
    letter-spacing: normal;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.filter-pill:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.filter-pill:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.filter-pill--active {
    border-color: rgba(var(--v-theme-secondary), 0.7);
    color: rgb(var(--v-theme-secondary));
}

.filter-pill--active:hover {
    background-color: rgba(var(--v-theme-secondary), 0.08);
    border-color: rgb(var(--v-theme-secondary));
}

.filter-pill--open {
    background-color: rgba(var(--v-theme-on-surface), 0.12);
    border-color: rgba(var(--v-theme-on-surface), 0.4);
}

.filter-pill__icon {
    opacity: 0.9;
}

.filter-pill__dot {
    opacity: 0.7;
    margin-left: 2px;
}

.filter-pill__count {
    font-weight: 600;
}

.filter-modal {
    width: min(95vw, 420px);
    border-radius: 16px;
    background-color: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    box-shadow: 0 16px 40px rgba(var(--v-theme-shadow), 0.5);
    color: rgba(var(--v-theme-on-surface), 0.92);
    overflow: hidden;
}

.filter-modal__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 16px 6px;
}

.filter-modal__title-row {
    display: inline-flex;
    align-items: center;
    gap: 10px;
}

.filter-modal__title {
    font-size: 1.0625rem;
    font-weight: 600;
}

.filter-modal__active-badge {
    display: inline-flex;
    align-items: center;
    height: 22px;
    padding: 0 9px;
    border-radius: 999px;
    border: 1px solid rgba(var(--v-theme-secondary), 0.55);
    color: rgb(var(--v-theme-secondary));
    font-size: 0.75rem;
    font-weight: 500;
}

.filter-modal__section {
    padding: 12px 16px;
}

.filter-modal__section + .filter-modal__section {
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.filter-modal__section-head {
    display: flex;
    align-items: baseline;
    gap: 8px;
    margin-bottom: 10px;
}

.filter-modal__section-title {
    font-size: 0.9375rem;
    font-weight: 600;
}

.filter-modal__section-value {
    margin-left: auto;
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.65);
    font-variant-numeric: tabular-nums;
}

.filter-modal__section-clear {
    background: transparent;
    border: 0;
    padding: 2px 4px;
    color: rgb(var(--v-theme-secondary));
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    border-radius: 4px;
}

.filter-modal__section-clear:hover {
    text-decoration: underline;
}

.filter-modal__section-clear:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.filter-slider {
    margin-top: 2px;
    padding-inline: 4px;
}

.filter-slider :deep(.v-slider-thumb__label) {
    background-color: rgb(var(--v-theme-secondary));
}

.filter-slider :deep(.v-slider__tick-label) {
    font-size: 0.6875rem;
    color: rgba(var(--v-theme-on-surface), 0.5);
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
}

.filter-slider :deep(.v-slider__tick:first-child .v-slider__tick-label) {
    transform: translateX(0);
    text-align: left;
}

.filter-slider :deep(.v-slider__tick:last-child .v-slider__tick-label) {
    transform: translateX(-100%);
    text-align: right;
}

.filter-segmented {
    display: flex;
    align-items: stretch;
    width: 100%;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.16);
    border-radius: 999px;
    overflow: hidden;
    background-color: transparent;
}

.filter-segmented__btn {
    flex: 1 1 0;
    min-width: 0;
    height: 36px;
    padding: 0 8px;
    background: transparent;
    border: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    color: rgba(var(--v-theme-on-surface), 0.82);
    font-size: 0.8125rem;
    font-weight: 500;
    letter-spacing: normal;
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.filter-segmented__btn:first-child {
    border-left: 0;
}

.filter-segmented__btn:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.filter-segmented__btn:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: -2px;
}

.filter-segmented__btn--active,
.filter-segmented__btn--active:hover {
    background-color: rgba(var(--v-theme-secondary), 0.18);
    color: rgb(var(--v-theme-secondary));
    font-weight: 600;
}

.filter-modal__footer {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px 14px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.filter-modal__reset {
    flex: 0 0 auto;
    height: 38px;
    padding: 0 16px;
    border-radius: 999px;
    background: transparent;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
    color: rgba(var(--v-theme-on-surface), 0.9);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, opacity 120ms ease;
}

.filter-modal__reset:hover:not(:disabled) {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.32);
}

.filter-modal__reset:disabled {
    opacity: 0.45;
    cursor: default;
}

.filter-modal__apply {
    flex: 1 1 auto;
    height: 38px;
    padding: 0 16px;
    border-radius: 999px;
    background-color: rgb(var(--v-theme-secondary));
    color: rgba(var(--v-theme-on-secondary), 0.87);
    border: 0;
    font-size: 0.9375rem;
    font-weight: 600;
    cursor: pointer;
    transition: filter 120ms ease;
}

.filter-modal__apply:hover {
    filter: brightness(1.06);
}

.filter-modal__apply:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}
</style>
