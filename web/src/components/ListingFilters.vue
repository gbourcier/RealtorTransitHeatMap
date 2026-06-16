<script setup lang="ts">
import { computed, ref } from "vue";
import type { ListingFiltersState } from "../composables/useListingFilters";
import type { SavedViewsState } from "../composables/useSavedViews";
import { LISTING_FILTER_BUILDING_TYPES } from "../constants/realtor";
import { formatCompactPrice } from "../utils/listingFormat";
import HeatSlider from "./HeatSlider.vue";

interface Props {
    state: ListingFiltersState;
    views: SavedViewsState;
    total: number;
}

const props = defineProps<Props>();
const emit = defineEmits<{
    close: [];
    error: [message: string];
    save: [];
}>();
const rootEl = ref<HTMLElement | null>(null);
defineExpose({ rootEl });

const PRICE_MIN = 300_000;
const PRICE_MAX = 1_000_000;
const PRICE_STEP = 25_000;
const PRICE_NO_MAX = PRICE_MAX + PRICE_STEP;
const priceTicks = [
    { at: 300_000, label: "$300k" },
    { at: 500_000, label: "$500k" },
    { at: 700_000, label: "$700k" },
    { at: 900_000, label: "$900k" },
    { at: PRICE_NO_MAX, label: "Any" },
];

const COMMUTE_MAX = 60;
const COMMUTE_STEP = 5;
const COMMUTE_NO_MAX = COMMUTE_MAX + COMMUTE_STEP;
const commuteTicks = [
    { at: 0, label: "0m" },
    { at: 30, label: "30m" },
    { at: COMMUTE_NO_MAX, label: "Any" },
];

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

const buildingTypeOptions = LISTING_FILTER_BUILDING_TYPES;

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

const buildingTypeLabel = computed(() => {
    if (props.state.buildingTypes.value.length === 0) return "Any";
    if (props.state.buildingTypes.value.length === 1) {
        const id = props.state.buildingTypes.value[0];
        return buildingTypeOptions.find((opt) => opt.id === id)?.label ?? "1 type";
    }
    return `${props.state.buildingTypes.value.length} types`;
});

function isBuildingTypeSelected(id: number): boolean {
    return props.state.buildingTypes.value.includes(id);
}

function toggleBuildingType(id: number): void {
    if (isBuildingTypeSelected(id)) {
        props.state.buildingTypes.value = props.state.buildingTypes.value.filter((v) => v !== id);
        return;
    }
    const allowedIds = buildingTypeOptions.map((opt) => opt.id);
    props.state.buildingTypes.value = [...props.state.buildingTypes.value, id]
        .filter((v) => allowedIds.includes(v))
        .sort((a, b) => allowedIds.indexOf(a) - allowedIds.indexOf(b));
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
    <div ref="rootEl" class="filter-drawer" role="dialog" aria-label="Filters">
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
                    @click="emit('save')"
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
                <div class="fp__sec-head">
                    <span class="fp__label">Building type</span>
                    <span
                        class="fp__val"
                        :class="{ 'fp__val--set': state.buildingTypes.value.length > 0 }"
                    >
                        {{ buildingTypeLabel }}
                    </span>
                </div>
                <div class="chips" aria-label="Building type">
                    <button
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': state.buildingTypes.value.length === 0 }"
                        :aria-pressed="state.buildingTypes.value.length === 0"
                        @click="state.buildingTypes.value = []"
                    >Any</button>
                    <button
                        v-for="opt in buildingTypeOptions"
                        :key="opt.id"
                        type="button"
                        class="chip"
                        :class="{ 'chip--on': isBuildingTypeSelected(opt.id) }"
                        :aria-pressed="isBuildingTypeSelected(opt.id)"
                        @click="toggleBuildingType(opt.id)"
                    >{{ opt.label }}</button>
                </div>
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
                </div>
            </section>

            <section class="fp__sec fp__sec--compact">
                <button
                    type="button"
                    class="fp__toggle"
                    :class="{ 'fp__toggle--on': state.includeExpired.value }"
                    :aria-pressed="state.includeExpired.value"
                    @click="state.includeExpired.value = !state.includeExpired.value"
                >
                    <span class="fp__check">
                        <v-icon v-if="state.includeExpired.value" size="14">mdi-check</v-icon>
                    </span>
                    Include expired
                </button>
            </section>

        </div>

        <footer class="fp__foot">
            <button class="fp__reset" @click="views.resetFilters()">Reset</button>
            <button class="fp__save" @click="emit('save')">
                <v-icon size="16">mdi-bookmark-outline</v-icon>
                Save
            </button>
            <button class="fp__apply" @click="emit('close')">
                Show <b>{{ total.toLocaleString() }}</b> result{{ total === 1 ? "" : "s" }}
            </button>
        </footer>

    </div>
</template>

<style scoped>
.filter-drawer {
    position: fixed;
    top: 76px;
    right: 76px;
    z-index: 1005;
    display: flex;
    flex-direction: column;
    width: min(440px, calc(100vw - 28px));
    max-height: 86vh;
    overflow: hidden;
    border-radius: 18px;
    background: #262925;
    border: 1px solid rgba(244, 241, 232, 0.12);
    box-shadow: 0 30px 70px -18px rgba(0, 0, 0, 0.85);
    color: #f4f1e8;
    font-family: Inter, system-ui, sans-serif;
}

.fp__head {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 18px 18px 14px;
}

.fp__title {
    font-size: 22px;
    font-weight: 700;
}

.fp__badge {
    display: inline-flex;
    align-items: center;
    padding: 3px 10px;
    border-radius: 999px;
    border: 1.5px solid #b6f24a;
    color: #b6f24a;
    font-size: 12.5px;
    font-weight: 700;
    white-space: nowrap;
}

.fp__x {
    margin-left: auto;
    width: 34px;
    height: 34px;
    flex: 0 0 auto;
    display: grid;
    place-items: center;
    border-radius: 9px;
    border: 1px solid rgba(244, 241, 232, 0.12);
    background: transparent;
    color: #f4f1e8;
    cursor: pointer;
}

.fp__x:hover,
.fp__reset:hover,
.fp__save:hover {
    background: #2a2d27;
}

.fp__x:focus-visible,
.chip:focus-visible,
.fp__reset:focus-visible,
.fp__save:focus-visible,
.fp__apply:focus-visible,
.fp__toggle:focus-visible {
    outline: 2px solid #6ccff6;
    outline-offset: 2px;
}

.fp__applied {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 9px;
    margin: 0 18px 18px;
    padding: 13px 15px;
    border-radius: 12px;
    background: rgba(182, 242, 74, 0.12);
    border: 1px solid rgba(182, 242, 74, 0.3);
}

.fp__applied__check,
.fp__applied__name {
    color: #b6f24a;
}

.fp__applied__text {
    min-width: 0;
    display: inline-flex;
    align-items: baseline;
    gap: 5px;
    overflow: hidden;
}

.fp__applied__name {
    overflow: hidden;
    font-size: 14.5px;
    font-weight: 700;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.fp__applied__state {
    flex: 0 0 auto;
    color: rgba(244, 241, 232, 0.52);
    font-size: 13px;
}

.fp__applied__state--dirty {
    color: #ffb454;
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
    font-size: 12.5px;
    font-weight: 600;
    white-space: nowrap;
}

.fp__applied__btn--primary {
    background: #b6f24a;
    border: 0;
    color: #172006;
}

.fp__applied__btn--ghost {
    background: transparent;
    border: 1px solid rgba(244, 241, 232, 0.22);
    color: #f4f1e8;
}

.fp__body {
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    gap: 18px;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 0 18px 18px;
}

.fp__body::-webkit-scrollbar {
    width: 10px;
}

.fp__body::-webkit-scrollbar-thumb {
    background: rgba(244, 241, 232, 0.14);
    border-radius: 999px;
    border: 3px solid #262925;
}

.hero {
    padding: 18px;
    border-radius: 14px;
    background: #181a17;
    border: 1px solid rgba(244, 241, 232, 0.12);
}

.hero__top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 16px;
}

.hero__kicker {
    max-width: 160px;
    color: #b6f24a;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.1em;
    line-height: 1.3;
    text-transform: uppercase;
}

.hero__read {
    color: #f4f1e8;
    font-size: 42px;
    font-weight: 800;
    line-height: 0.9;
    font-variant-numeric: tabular-nums;
}

.hero__read small {
    margin-left: 4px;
    color: rgba(244, 241, 232, 0.52);
    font-size: 17px;
    font-weight: 600;
}

.hero__counts {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
}

.hero__total {
    color: #b6f24a;
    font-size: 15px;
    font-weight: 800;
    font-variant-numeric: tabular-nums;
}

.hero__total small {
    margin-left: 5px;
    color: rgba(244, 241, 232, 0.52);
    font-size: 13px;
    font-weight: 500;
}

.fp__sec {
    padding: 0;
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
    font-size: 16px;
    font-weight: 700;
}

.fp__val {
    margin-left: auto;
    color: rgba(244, 241, 232, 0.52);
    font-size: 15px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
}

.fp__val--set {
    color: #b6f24a;
}

.chips {
    display: flex;
    flex-wrap: wrap;
    gap: 9px;
}

.chip {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 48px;
    height: 40px;
    padding: 0 16px;
    border-radius: 999px;
    border: 1px solid rgba(244, 241, 232, 0.12);
    background: transparent;
    color: #f4f1e8;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
}

.chip:hover {
    border-color: rgba(244, 241, 232, 0.22);
}

.chip--on,
.chip--on:hover {
    background: rgba(182, 242, 74, 0.1);
    border: 1.5px solid #b6f24a;
    color: #b6f24a;
}

.fp__toggle {
    display: inline-flex;
    align-items: center;
    gap: 11px;
    height: 46px;
    padding: 0 16px 0 13px;
    border-radius: 12px;
    border: 1px solid rgba(244, 241, 232, 0.12);
    background: transparent;
    color: #f4f1e8;
    font: inherit;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
}

.fp__toggle:hover {
    border-color: rgba(244, 241, 232, 0.22);
}

.fp__toggle--on {
    border-color: #b6f24a;
}

.fp__check {
    display: grid;
    place-items: center;
    width: 20px;
    height: 20px;
    border-radius: 6px;
    border: 1.5px solid rgba(244, 241, 232, 0.22);
}

.fp__toggle--on .fp__check {
    background: #b6f24a;
    border-color: #b6f24a;
    color: #172006;
}

.fp__foot {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 14px 18px;
    border-top: 1px solid rgba(244, 241, 232, 0.12);
    background: #262925;
}

.fp__reset,
.fp__save {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 46px;
    padding: 0 18px;
    border-radius: 999px;
    border: 1px solid rgba(244, 241, 232, 0.22);
    background: transparent;
    color: #f4f1e8;
    font-size: 15px;
    font-weight: 600;
    cursor: pointer;
}

.fp__apply {
    flex: 1 1 auto;
    height: 46px;
    padding: 0 22px;
    border-radius: 999px;
    background: #b6f24a;
    color: #172006;
    border: 0;
    font-size: 15.5px;
    font-weight: 700;
    white-space: nowrap;
    cursor: pointer;
}

.fp__apply:hover {
    background: #c6fb60;
}

@media (max-width: 899px) {
    .filter-drawer {
        top: auto;
        right: 11px;
        bottom: 0;
        left: 11px;
        width: auto;
        max-height: 88dvh;
        border-radius: 22px 22px 0 0;
    }

    .fp__applied {
        align-items: flex-start;
    }

    .fp__applied__actions {
        display: none;
    }

    .fp__foot {
        display: grid;
        grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
        gap: 9px;
        padding: 12px;
    }

    .fp__reset,
    .fp__save,
    .fp__apply {
        justify-content: center;
        min-width: 0;
        width: 100%;
        height: 44px;
    }

    .fp__reset,
    .fp__save {
        padding: 0 12px;
    }

    .fp__apply {
        grid-column: 1 / -1;
        padding: 0 16px;
    }
}

</style>
