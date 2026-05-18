<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from "vue";
import { useDisplay } from "vuetify";
import {
    listListings,
    type Listing,
    type SortBy,
    type SortDir,
} from "../api/listings";
import ListingsMap from "../components/ListingsMap.vue";

const { mdAndUp } = useDisplay();
const viewMode = ref<"list" | "map">("map");
const drawerOpen = ref(true);

const items = ref<Listing[]>([]);
const total = ref(0);
const mapCount = ref(0);

const limit = ref(50);
const loading = ref(false);
const error = ref<string | null>(null);
let loadGen = 0;
const sortBy = ref<SortBy>("first_seen_at");
const sortDir = ref<SortDir>("desc");

const maxPrice = ref<number | null>(null);
const maxCommuteSec = ref<number | null>(null);
const newWithinDays = ref<number | null>(null);
const minBedrooms = ref<number | null>(1);
const minBathrooms = ref<number | null>(1);
const minInteriorAreaSqft = ref<number | null>(null);

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

const priceSlider = computed({
    get: () => maxPrice.value ?? PRICE_NO_MAX,
    set: (v: number) => {
        maxPrice.value = v >= PRICE_NO_MAX ? null : v;
        applyFilters();
    },
});

const commuteSliderMin = computed({
    get: () =>
        maxCommuteSec.value == null ? COMMUTE_NO_MAX : Math.round(maxCommuteSec.value / 60),
    set: (v: number) => {
        maxCommuteSec.value = v >= COMMUTE_NO_MAX ? null : v * 60;
        applyFilters();
    },
});

const sqftSlider = computed({
    get: () => minInteriorAreaSqft.value ?? SQFT_MIN,
    set: (v: number) => {
        minInteriorAreaSqft.value = v <= SQFT_MIN ? null : v;
        applyFilters();
    },
});

const filterMenuOpen = ref(false);

const activeFilterCount = computed(() => {
    let n = 0;
    if (maxPrice.value != null) n++;
    if (maxCommuteSec.value != null) n++;
    if (newWithinDays.value != null) n++;
    if (minBedrooms.value != null && minBedrooms.value > 1) n++;
    if (minBathrooms.value != null && minBathrooms.value > 1) n++;
    if (minInteriorAreaSqft.value != null) n++;
    return n;
});

function formatCompactPrice(price: number): string {
    if (price >= 1_000_000) {
        const m = price / 1_000_000;
        return `$${m % 1 === 0 ? m.toFixed(0) : m.toFixed(1)}M`;
    }
    return `$${Math.round(price / 1000)}k`;
}

function applyFilters() {
    loadInitial();
}

function setMaxPrice(value: number | null) {
    maxPrice.value = value;
    applyFilters();
}

function setMaxCommute(minutes: number | null) {
    maxCommuteSec.value = minutes == null ? null : minutes * 60;
    applyFilters();
}

function setRecency(days: number | null) {
    newWithinDays.value = days;
    applyFilters();
}

function setMinBedrooms(value: number | null) {
    minBedrooms.value = value;
    applyFilters();
}

function setMinBathrooms(value: number | null) {
    minBathrooms.value = value;
    applyFilters();
}

function setMinInteriorAreaSqft(value: number | null) {
    minInteriorAreaSqft.value = value;
    applyFilters();
}

function clearAllFilters() {
    maxPrice.value = null;
    maxCommuteSec.value = null;
    newWithinDays.value = null;
    minBedrooms.value = 1;
    minBathrooms.value = 1;
    minInteriorAreaSqft.value = null;
    applyFilters();
}

const hasMore = computed(() =>
    items.value.length === 0 ? true : items.value.length < total.value,
);

async function loadInitial(): Promise<void> {
    const gen = ++loadGen;
    items.value = [];
    total.value = 0;
    await loadMore(gen);
}

let inFlight: Promise<void> | null = null;

function loadMore(gen: number = loadGen): Promise<void> {
    if (gen !== loadGen) return Promise.resolve();
    if (inFlight) return inFlight.then(() => loadMore(gen));
    if (items.value.length > 0 && items.value.length >= total.value) {
        return Promise.resolve();
    }
    loading.value = true;
    error.value = null;
    inFlight = (async () => {
        try {
            const res = await listListings({
                limit: limit.value,
                offset: items.value.length,
                sortBy: sortBy.value,
                sortDir: sortDir.value,
                ...(maxPrice.value != null && { maxPrice: maxPrice.value }),
                ...(maxCommuteSec.value != null && { maxCommuteSec: maxCommuteSec.value }),
                ...(newWithinDays.value != null && { newWithinDays: newWithinDays.value }),
                ...(minBedrooms.value != null && minBedrooms.value > 1 && { minBedrooms: minBedrooms.value }),
                ...(minBathrooms.value != null && minBathrooms.value > 1 && { minBathrooms: minBathrooms.value }),
                ...(minInteriorAreaSqft.value != null && { minInteriorAreaSqft: minInteriorAreaSqft.value }),
            });
            if (gen !== loadGen) return;
            items.value = [...items.value, ...res.items];
            total.value = res.total;
        } catch (e: any) {
            if (gen === loadGen) {
                error.value =
                    e?.response?.data?.error ?? e?.message ?? "failed to load listings";
            }
        } finally {
            if (gen === loadGen) loading.value = false;
            inFlight = null;
        }
    })();
    return inFlight;
}

interface SortOption {
    value: SortBy;
    label: string;
    defaultDir: SortDir;
}

const sortOptions: SortOption[] = [
    {
        value: "first_seen_at",
        label: "Newest",
        defaultDir: "desc",
    },
    {
        value: "price",
        label: "Price",
        defaultDir: "asc",
    },
    {
        value: "commute_time",
        label: "Commute",
        defaultDir: "asc",
    },
];

function selectSort(opt: SortOption) {
    if (sortBy.value === opt.value) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    } else {
        sortBy.value = opt.value;
        sortDir.value = opt.defaultDir;
    }
    loadInitial();
}

function formatPrice(price: number | null): string {
    if (price == null) return "—";
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
        maximumFractionDigits: 0,
    }).format(price);
}

function daysSince(unix: number): number {
    const date = new Date(unix * 1000);
    const now = new Date();
    const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return Math.round((startOfToday.getTime() - startOfDate.getTime()) / 86400000);
}

function formatDate(unix: number): string {
    const diff = daysSince(unix);
    if (diff === 0) return "today";
    if (diff === 1) return "yesterday";
    const date = new Date(unix * 1000);
    const dd = String(date.getDate()).padStart(2, "0");
    const mm = String(date.getMonth() + 1).padStart(2, "0");
    const yyyy = date.getFullYear();
    return `${dd}-${mm}-${yyyy}`;
}

function isNew(unix: number): boolean {
    return daysSince(unix) === 0;
}

function formatCommute(seconds: number | null): string {
    if (seconds == null) return "—";
    return `${Math.round(seconds / 60)} min`;
}

function commuteColor(seconds: number | null): string {
    if (seconds == null) return "rgba(var(--v-theme-on-surface), 0.35)";
    const minutes = seconds / 60;
    if (minutes < 30) return "rgb(var(--v-theme-success))";
    if (minutes <= 60) return "rgb(var(--v-theme-warning))";
    return "rgb(var(--v-theme-error))";
}

function commuteMapUrl(address: string | null): string | null {
    if (!address) return null;
    const params = new URLSearchParams({
        saddr: address,
        daddr: "McGill Station, Montreal, QC",
        dirflg: "r",
        ttype: "arr",
    });
    return `https://www.google.com/maps?${params.toString()}`;
}

const mapRef = ref<InstanceType<typeof ListingsMap> | null>(null);
const selectedKey = ref<string | null>(null);

function listingKey(item: Listing): string {
    return `${item.board}-${item.mls}`;
}

function openListing(item: Listing): void {
    if (!item.slug) return;
    window.open(item.slug, "_blank", "noopener,noreferrer");
}

function focusListingOnMap(item: Listing): void {
    selectedKey.value = listingKey(item);
    mapRef.value?.focusListing(item.board, item.mls);
}

function highlightListingOnMap(item: Listing): void {
    mapRef.value?.highlightListing(item.board, item.mls);
}

function clearMapHighlight(): void {
    mapRef.value?.clearHighlight();
}

function parseAddress(raw: string | null | undefined): {
    street: string;
    locality: string;
} {
    if (!raw) return { street: "—", locality: "" };
    const parts = raw.split("|").map((s) => s.trim()).filter(Boolean);
    const street = parts[0] ?? raw;
    const rest = parts.slice(1).join(", ");
    const parenMatch = rest.match(/\(([^)]+)\)/);
    const locality = parenMatch
        ? parenMatch[1].trim()
        : rest
            .replace(/\s+[A-Z]\d[A-Z]\s?\d[A-Z]\d\s*$/i, "")
            .replace(/,\s*(Qu[eé]bec|QC)\b\.?/gi, "")
            .replace(/,\s*$/, "")
            .trim();
    return { street, locality };
}

const teleportReady = ref(false);
const sidePanelBodyEl = ref<HTMLElement | null>(null);
const panelSentinelEl = ref<HTMLElement | null>(null);
const mobileSentinelEl = ref<HTMLElement | null>(null);
let panelObserver: IntersectionObserver | null = null;
let mobileObserver: IntersectionObserver | null = null;

function makeObserver(
    sentinel: HTMLElement | null,
    root: HTMLElement | null,
): IntersectionObserver | null {
    if (!sentinel) return null;
    const obs = new IntersectionObserver(
        (entries) => {
            if (entries.some((e) => e.isIntersecting)) loadMore();
        },
        { root, rootMargin: "200px" },
    );
    obs.observe(sentinel);
    return obs;
}

watch(
    [panelSentinelEl, sidePanelBodyEl],
    ([sentinel, body]) => {
        panelObserver?.disconnect();
        panelObserver = makeObserver(sentinel, body);
    },
    { flush: "post" },
);

watch(
    mobileSentinelEl,
    (sentinel) => {
        mobileObserver?.disconnect();
        mobileObserver = makeObserver(sentinel, null);
    },
    { flush: "post" },
);

let prevHtmlOverflow = "";
let prevBodyOverflow = "";

onMounted(() => {
    teleportReady.value = true;
    prevHtmlOverflow = document.documentElement.style.overflow;
    prevBodyOverflow = document.body.style.overflow;
    document.documentElement.style.overflow = "hidden";
    document.body.style.overflow = "hidden";
    loadInitial();
});

onBeforeUnmount(() => {
    panelObserver?.disconnect();
    mobileObserver?.disconnect();
    document.documentElement.style.overflow = prevHtmlOverflow;
    document.body.style.overflow = prevBodyOverflow;
});
</script>

<template>
    <Teleport to="#header-filters-slot" :disabled="!teleportReady">
        <v-menu v-model="filterMenuOpen" :close-on-content-click="false" location="bottom end" offset="18"
            transition="scale-transition">
            <template #activator="{ props }">
                <button v-bind="props" type="button" class="filter-pill"
                    :class="{ 'filter-pill--active': activeFilterCount > 0 }">
                    <v-icon size="16" class="filter-pill__icon">mdi-tune-variant</v-icon>
                    <span class="filter-pill__label">Filters</span>
                    <template v-if="activeFilterCount > 0">
                        <span class="filter-pill__dot" aria-hidden="true">•</span>
                        <span class="filter-pill__count">{{ activeFilterCount }}</span>
                    </template>
                </button>
            </template>
            <div class="filter-modal" role="dialog" aria-label="Filters">
                <header class="filter-modal__header">
                    <div class="filter-modal__title-row">
                        <span class="filter-modal__title">Filters</span>
                        <span v-if="activeFilterCount > 0" class="filter-modal__active-badge">
                            {{ activeFilterCount }} active
                        </span>
                    </div>
                    <button type="button" class="filter-modal__close" aria-label="Close filters"
                        @click="filterMenuOpen = false">
                        <v-icon size="20">mdi-close</v-icon>
                    </button>
                </header>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Max price</span>
                        <span class="filter-modal__section-value">
                            {{ maxPrice == null ? "Any" : formatCompactPrice(maxPrice) }}
                        </span>
                        <button v-if="maxPrice != null" type="button" class="filter-modal__section-clear"
                            @click="setMaxPrice(null)">Clear</button>
                    </div>
                    <v-slider v-model="priceSlider" :min="PRICE_MIN" :max="PRICE_NO_MAX" :step="PRICE_STEP"
                        :ticks="priceTicks" show-ticks="always" tick-size="3" color="secondary"
                        track-color="rgba(var(--v-theme-on-surface), 0.16)" hide-details density="compact" class="filter-slider" />
                </section>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Max commute</span>
                        <span class="filter-modal__section-value">
                            {{ maxCommuteSec == null ? "Any" : `${Math.round(maxCommuteSec / 60)} min` }}
                        </span>
                        <button v-if="maxCommuteSec != null" type="button" class="filter-modal__section-clear"
                            @click="setMaxCommute(null)">Clear</button>
                    </div>
                    <v-slider v-model="commuteSliderMin" :min="COMMUTE_MIN" :max="COMMUTE_NO_MAX" :step="COMMUTE_STEP"
                        :ticks="commuteTicks" show-ticks="always" tick-size="3" color="secondary"
                        track-color="rgba(var(--v-theme-on-surface), 0.16)" hide-details density="compact" class="filter-slider" />
                </section>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Min interior space</span>
                        <span class="filter-modal__section-value">
                            {{ minInteriorAreaSqft == null ? "Any" : `${minInteriorAreaSqft.toLocaleString()} sqft` }}
                        </span>
                        <button v-if="minInteriorAreaSqft != null" type="button" class="filter-modal__section-clear"
                            @click="setMinInteriorAreaSqft(null)">Clear</button>
                    </div>
                    <v-slider v-model="sqftSlider" :min="SQFT_MIN" :max="SQFT_MAX" :step="SQFT_STEP" :ticks="sqftTicks"
                        show-ticks="always" tick-size="3" color="secondary" track-color="rgba(var(--v-theme-on-surface), 0.16)"
                        hide-details density="compact" class="filter-slider" />
                </section>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Bedrooms</span>
                    </div>
                    <div class="filter-segmented" role="radiogroup" aria-label="Minimum bedrooms">
                        <button v-for="b in bedroomOptions" :key="b" type="button" class="filter-segmented__btn"
                            :class="{ 'filter-segmented__btn--active': minBedrooms === b }" role="radio"
                            :aria-checked="minBedrooms === b" @click="setMinBedrooms(b)">
                            {{ b }}+
                        </button>
                    </div>
                </section>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Bathrooms</span>
                    </div>
                    <div class="filter-segmented" role="radiogroup" aria-label="Minimum bathrooms">
                        <button v-for="b in bathroomOptions" :key="b" type="button" class="filter-segmented__btn"
                            :class="{ 'filter-segmented__btn--active': minBathrooms === b }" role="radio"
                            :aria-checked="minBathrooms === b" @click="setMinBathrooms(b)">
                            {{ b }}+
                        </button>
                    </div>
                </section>

                <section class="filter-modal__section">
                    <div class="filter-modal__section-head">
                        <span class="filter-modal__section-title">Recency</span>
                    </div>
                    <div class="filter-segmented" role="radiogroup" aria-label="Listing recency">
                        <button v-for="opt in recencyOptions" :key="opt.label" type="button"
                            class="filter-segmented__btn"
                            :class="{ 'filter-segmented__btn--active': newWithinDays === opt.days }" role="radio"
                            :aria-checked="newWithinDays === opt.days" @click="setRecency(opt.days)">
                            {{ opt.label }}
                        </button>
                    </div>
                </section>

                <footer class="filter-modal__footer">
                    <button type="button" class="filter-modal__reset" :disabled="activeFilterCount === 0"
                        @click="clearAllFilters">Reset</button>
                    <button type="button" class="filter-modal__apply" @click="filterMenuOpen = false">
                        Show {{ total.toLocaleString() }} listing{{ total === 1 ? "" : "s" }}
                    </button>
                </footer>
            </div>
        </v-menu>

    </Teleport>

    <Teleport to="#header-actions-slot" :disabled="!teleportReady">
        <v-btn v-if="mdAndUp" icon variant="text" size="small" :active="drawerOpen"
            :aria-label="drawerOpen ? 'Hide results panel' : 'Show results panel'" @click="drawerOpen = !drawerOpen">
            <v-icon size="22">mdi-dock-right</v-icon>
        </v-btn>
    </Teleport>

    <div v-show="mdAndUp || viewMode === 'map'" class="map-fullbleed"
        :class="{ 'map-fullbleed--with-panel': mdAndUp && drawerOpen }">
        <ListingsMap ref="mapRef" class="map-fullbleed__map" :max-price="maxPrice" :max-commute-sec="maxCommuteSec"
            :new-within-days="newWithinDays" :min-bedrooms="minBedrooms" :min-bathrooms="minBathrooms"
            :min-interior-area-sqft="minInteriorAreaSqft" @update:count="mapCount = $event" />
        <aside v-if="mdAndUp && drawerOpen" class="listings-side-panel">
            <div class="list-toolbar">
                <div class="list-toolbar__row">
                    <div class="list-toolbar__count">
                        <span class="list-toolbar__count-num">{{ total.toLocaleString() }}</span>
                        <span class="list-toolbar__count-label">listing{{ total === 1 ? "" : "s" }}</span>
                    </div>
                    <div class="sort-tabs" role="tablist" aria-label="Sort listings">
                        <button v-for="opt in sortOptions" :key="opt.value" type="button" role="tab"
                            class="sort-tabs__tab" :class="{ 'sort-tabs__tab--active': sortBy === opt.value }"
                            :aria-selected="sortBy === opt.value" @click="selectSort(opt)">
                            <span class="sort-tabs__label">{{ opt.label }}</span>
                            <v-icon v-if="sortBy === opt.value" size="14" class="sort-tabs__dir">
                                {{ sortDir === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down' }}
                            </v-icon>
                        </button>
                    </div>
                </div>
            </div>
            <div ref="sidePanelBodyEl" class="listings-side-panel__body">
                <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{ error }}</v-alert>

                <div v-if="loading && items.length === 0" class="text-center py-8">
                    <v-progress-circular indeterminate />
                </div>

                <template v-else-if="items.length > 0">
                    <div class="listing-cards listing-cards--panel">
                        <div v-for="item in items" :key="`p-${item.board}-${item.mls}`" role="button" tabindex="0"
                            class="listing-card listing-card--interactive"
                            :class="{ 'listing-card--selected': selectedKey === listingKey(item) }"
                            @click="focusListingOnMap(item)" @keydown.enter.prevent="focusListingOnMap(item)"
                            @keydown.space.prevent="focusListingOnMap(item)" @mouseenter="highlightListingOnMap(item)"
                            @mouseleave="clearMapHighlight" @focus="highlightListingOnMap(item)"
                            @blur="clearMapHighlight">
                            <div class="listing-card__top">
                                <span class="listing-card__price">{{
                                    formatPrice(item.currentPrice)
                                }}</span>
                                <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary" variant="flat"
                                    class="listing-card__new">new</v-chip>
                            </div>
                            <div class="listing-card__street">
                                {{ parseAddress(item.address).street }}
                            </div>
                            <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
                                {{ parseAddress(item.address).locality }}
                            </div>
                            <div class="listing-card__meta">
                                <a v-if="item.commuteSecondsDowntown != null && item.address"
                                    :href="commuteMapUrl(item.address) ?? '#'" target="_blank" rel="noopener noreferrer"
                                    class="listing-card__commute listing-card__commute--link" @click.stop>
                                    <span class="listing-card__commute-dot"
                                        :style="{ background: commuteColor(item.commuteSecondsDowntown) }" />
                                    <span>{{ formatCommute(item.commuteSecondsDowntown) }} downtown</span>
                                </a>
                                <span v-else class="listing-card__commute listing-card__commute--muted">
                                    <span class="listing-card__commute-dot"
                                        :style="{ background: commuteColor(null) }" />
                                    —
                                </span>
                                <span class="listing-card__seen">{{
                                    formatDate(item.firstSeenAt)
                                }}</span>
                            </div>
                        </div>
                    </div>
                </template>

                <div v-else class="text-medium-emphasis text-center py-8">
                    No listings found.
                </div>

                <div v-if="hasMore && items.length > 0" ref="panelSentinelEl" class="listings-side-panel__sentinel">
                    <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                </div>
            </div>
        </aside>
    </div>

    <div v-if="!mdAndUp && viewMode === 'list'" class="listings-mobile">
        <div class="list-toolbar list-toolbar--mobile">
            <div class="list-toolbar__row">
                <div class="list-toolbar__count">
                    <span class="list-toolbar__count-num">{{ total.toLocaleString() }}</span>
                    <span class="list-toolbar__count-label">listing{{ total === 1 ? "" : "s" }}</span>
                </div>
                <div class="sort-tabs" role="tablist" aria-label="Sort listings">
                    <button v-for="opt in sortOptions" :key="opt.value" type="button" role="tab" class="sort-tabs__tab"
                        :class="{ 'sort-tabs__tab--active': sortBy === opt.value }"
                        :aria-selected="sortBy === opt.value" @click="selectSort(opt)">
                        <span class="sort-tabs__label">{{ opt.label }}</span>
                        <v-icon v-if="sortBy === opt.value" size="14" class="sort-tabs__dir">
                            {{ sortDir === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down' }}
                        </v-icon>
                    </button>
                </div>
            </div>
        </div>

        <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{
            error
        }}</v-alert>

        <div v-if="loading && items.length === 0" class="text-center py-8">
            <v-progress-circular indeterminate />
        </div>

        <template v-else-if="items.length > 0">
            <div class="listing-cards listing-cards--mobile">
                <div v-for="item in items" :key="`m-${item.board}-${item.mls}`" role="link" tabindex="0"
                    class="listing-card" @click="openListing(item)" @keydown.enter.prevent="openListing(item)"
                    @keydown.space.prevent="openListing(item)">
                    <div class="listing-card__top">
                        <span class="listing-card__price">{{
                            formatPrice(item.currentPrice)
                        }}</span>
                        <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary" variant="tonal"
                            class="listing-card__new">new</v-chip>
                    </div>
                    <div class="listing-card__street">
                        {{ parseAddress(item.address).street }}
                    </div>
                    <div v-if="parseAddress(item.address).locality" class="listing-card__locality">
                        {{ parseAddress(item.address).locality }}
                    </div>
                    <div class="listing-card__meta">
                        <a v-if="item.commuteSecondsDowntown != null && item.address"
                            :href="commuteMapUrl(item.address) ?? '#'" target="_blank" rel="noopener noreferrer"
                            class="listing-card__commute listing-card__commute--link" @click.stop>
                            <span class="listing-card__commute-dot"
                                :style="{ background: commuteColor(item.commuteSecondsDowntown) }" />
                            <span>{{ formatCommute(item.commuteSecondsDowntown) }} downtown</span>
                        </a>
                        <span v-else class="listing-card__commute listing-card__commute--muted">
                            <span class="listing-card__commute-dot" :style="{ background: commuteColor(null) }" />
                            —
                        </span>
                        <span class="listing-card__seen">{{
                            formatDate(item.firstSeenAt)
                        }}</span>
                    </div>
                </div>

                <div v-if="hasMore && items.length > 0" ref="mobileSentinelEl" class="listing-cards__sentinel">
                    <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                </div>
            </div>
        </template>

        <div v-else class="text-medium-emphasis text-center py-8">
            No listings found.
        </div>
    </div>

    <div v-if="!mdAndUp && viewMode === 'map'" class="mobile-legend" aria-label="Transit commute time legend">
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--fast" />
            &lt; 30
        </span>
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--mid" />
            30–60
        </span>
        <span class="mobile-legend__item">
            <span class="mobile-legend__dot mobile-legend__dot--slow" />
            &gt; 60 min
        </span>
    </div>

    <div v-if="!mdAndUp" class="mobile-view-toggle" role="tablist" aria-label="View mode">
        <button type="button" class="mobile-view-toggle__btn"
            :class="{ 'mobile-view-toggle__btn--active': viewMode === 'map' }" role="tab"
            :aria-selected="viewMode === 'map'" aria-label="Map view" @click="viewMode = 'map'">
            <v-icon size="20">mdi-map-outline</v-icon>
        </button>
        <button type="button" class="mobile-view-toggle__btn"
            :class="{ 'mobile-view-toggle__btn--active': viewMode === 'list' }" role="tab"
            :aria-selected="viewMode === 'list'" aria-label="List view" @click="viewMode = 'list'">
            <v-icon size="20">mdi-format-list-bulleted</v-icon>
        </button>
    </div>
</template>

<style scoped>
.filter-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    height: 32px;
    padding: 0 14px;
    border-radius: 999px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.22);
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.88);
    font-size: 0.8125rem;
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

.filter-modal__close {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 30px;
    height: 30px;
    border-radius: 50%;
    background: transparent;
    border: 0;
    color: rgba(var(--v-theme-on-surface), 0.7);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.filter-modal__close:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.filter-modal__close:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.filter-modal__section {
    padding: 12px 16px;
}

.filter-modal__section+.filter-modal__section {
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

.listings-mobile {
    height: calc(100dvh - 56px);
    overflow-y: auto;
    padding-bottom: calc(80px + env(safe-area-inset-bottom, 0px));
}

.map-fullbleed {
    height: calc(100dvh - 56px);
    width: 100%;
    display: flex;
}

.map-fullbleed__map {
    flex: 1 1 auto;
    min-height: 0;
    min-width: 0;
}

.listings-side-panel {
    flex: 0 0 360px;
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    background-color: rgb(var(--v-theme-surface));
}

.listings-side-panel__body {
    flex: 1 1 auto;
    overflow-y: auto;
    min-height: 0;
}

.listings-side-panel__sentinel,
.listing-cards__sentinel {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 36px;
    padding: 8px 0;
}

.listing-cards--panel {
    padding: 10px;
}

.list-toolbar {
    display: flex;
    flex-direction: column;
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    flex: 0 0 auto;
}

.list-toolbar--mobile {
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.list-toolbar__row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-width: 0;
}

.list-toolbar__count {
    display: flex;
    flex-direction: column;
    line-height: 1;
}

.list-toolbar__count-num {
    font-size: 1rem;
    font-weight: 700;
    letter-spacing: -0.01em;
}

.list-toolbar__count-label {
    font-size: 0.6875rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin-top: 2px;
}

.sort-tabs {
    display: flex;
    align-items: stretch;
    flex: 0 1 auto;
    min-width: 0;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 999px;
    overflow: hidden;
    background-color: transparent;
}

.sort-tabs__tab {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 0;
    height: 30px;
    padding: 0 12px;
    background: transparent;
    border: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 0;
    cursor: pointer;
    color: rgba(var(--v-theme-on-surface), 0.7);
    font-size: 0.8125rem;
    font-weight: 500;
    letter-spacing: normal;
    transition: background-color 120ms ease, color 120ms ease;
}

.sort-tabs__tab:first-child {
    border-left: 0;
}

.sort-tabs__tab:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.sort-tabs__tab:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: -2px;
}

.sort-tabs__tab--active {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.98);
    font-weight: 600;
}

.sort-tabs__label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.sort-tabs__dir {
    flex-shrink: 0;
    opacity: 0.9;
}

.mobile-legend {
    position: fixed;
    left: 12px;
    bottom: calc(20px + env(safe-area-inset-bottom, 0px));
    z-index: 1000;
    display: inline-flex;
    align-items: center;
    gap: 14px;
    height: 40px;
    padding: 0 16px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-surface), 0.78);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 6px 18px rgba(var(--v-theme-shadow), 0.45);
    color: rgba(var(--v-theme-on-surface), 0.92);
    font-size: 0.75rem;
    white-space: nowrap;
}

.mobile-legend__item {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.mobile-legend__dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.mobile-legend__dot--fast {
    background-color: rgb(var(--v-theme-success));
}

.mobile-legend__dot--mid {
    background-color: rgb(var(--v-theme-warning));
}

.mobile-legend__dot--slow {
    background-color: rgb(var(--v-theme-error));
}

.mobile-view-toggle {
    position: fixed;
    right: 12px;
    bottom: calc(20px + env(safe-area-inset-bottom, 0px));
    z-index: 1000;
    display: inline-flex;
    align-items: center;
    height: 40px;
    padding: 4px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-surface), 0.78);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 6px 18px rgba(var(--v-theme-shadow), 0.45);
}

.mobile-view-toggle__btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 44px;
    height: 32px;
    padding: 0;
    background: transparent;
    border: 0;
    border-radius: 999px;
    color: rgba(var(--v-theme-on-surface), 0.7);
    cursor: pointer;
    transition: background-color 120ms ease, color 120ms ease;
}

.mobile-view-toggle__btn:hover {
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.mobile-view-toggle__btn:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.mobile-view-toggle__btn--active {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
    color: rgb(var(--v-theme-on-surface));
}

.listing-cards {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 10px;
}

.listing-cards--mobile {
    gap: 14px;
    padding: 10px 12px 12px;
}

.listing-card {
    position: relative;
    display: block;
    text-decoration: none;
    color: inherit;
    background-color: rgba(var(--v-theme-on-surface), 0.03);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 14px;
    padding: 14px 14px 12px;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-cards--mobile .listing-card {
    padding: 18px 18px 16px;
}

.listing-cards--mobile .listing-card__price {
    font-size: 1.6rem;
    font-weight: 700;
}

.listing-cards--mobile .listing-card__street {
    font-size: 1rem;
    font-weight: 500;
}

.listing-cards--mobile .listing-card__meta {
    border-top: 0;
    padding-top: 14px;
    margin-top: 10px;
}

.listing-cards--mobile .listing-card__new {
    top: 16px;
    right: 16px;
}

.listing-card:active {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.listing-card:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card--interactive:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
}

.listing-card--selected,
.listing-card--selected:hover {
    background-color: rgba(var(--v-theme-secondary), 0.12);
    border-color: rgba(var(--v-theme-secondary), 0.55);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-secondary), 0.45);
}

.listing-card__top {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    padding-right: 56px;
}

.listing-card__price {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.2;
}

.listing-card__new {
    position: absolute;
    top: 12px;
    right: 12px;
    margin-left: 0;
}

.listing-card__street {
    font-size: 0.95rem;
    line-height: 1.3;
}

.listing-card__locality {
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.65);
    margin-top: 2px;
}

.listing-card__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    font-size: 0.8125rem;
}

.listing-card__commute {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    color: rgba(var(--v-theme-on-surface), 0.85);
    font-weight: 500;
    font-size: 0.8125rem;
}

.listing-card__commute--link {
    text-decoration: none;
    padding: 6px 12px;
    margin: -6px 0 -6px -4px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-on-surface), 0.06);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    transition: background-color 120ms ease, border-color 120ms ease;
}

.listing-card__commute--link:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.1);
    border-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:active {
    background-color: rgba(var(--v-theme-on-surface), 0.14);
}

.listing-card__commute--link:focus-visible {
    outline: 2px solid rgb(var(--v-theme-secondary));
    outline-offset: 2px;
}

.listing-card__commute--muted {
    color: rgba(var(--v-theme-on-surface), 0.5);
    font-weight: 400;
}

.listing-card__commute-dot {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.45);
    font-size: 0.75rem;
}
</style>
