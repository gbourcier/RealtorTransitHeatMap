<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from "vue";
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

function toggleViewMode() {
    viewMode.value = viewMode.value === "list" ? "map" : "list";
}
const limit = ref(50);
const loading = ref(false);
const error = ref<string | null>(null);
let loadGen = 0;
const sortBy = ref<SortBy>("first_seen_at");
const sortDir = ref<SortDir>("desc");

const maxPrice = ref<number | null>(null);
const maxCommuteSec = ref<number | null>(null);
const newWithinDays = ref<number | null>(null);

const priceOptions = [400000, 500000, 600000, 700000, 800000, 1000000, 1500000, 2000000];
const commuteOptions = [15, 30, 45, 60, 90];

type FilterTarget =
    | "#drawer-filters-slot"
    | "#mobile-filters-slot"
    | "#header-filters-slot";
const filterTarget = ref<FilterTarget>("#header-filters-slot");
watch(
    [mdAndUp, drawerOpen, viewMode],
    async ([md, open, mode]) => {
        let desired: FilterTarget;
        if (md) {
            desired = open ? "#drawer-filters-slot" : "#header-filters-slot";
        } else {
            desired = mode === "list" ? "#mobile-filters-slot" : "#header-filters-slot";
        }
        await nextTick();
        filterTarget.value = desired;
    },
    { immediate: true },
);

const activeFilterCount = computed(() => {
    let n = 0;
    if (maxPrice.value != null) n++;
    if (maxCommuteSec.value != null) n++;
    if (newWithinDays.value != null) n++;
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

function toggleNewOnly() {
    newWithinDays.value = newWithinDays.value == null ? 1 : null;
    applyFilters();
}

function clearAllFilters() {
    maxPrice.value = null;
    maxCommuteSec.value = null;
    newWithinDays.value = null;
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

const sortOptions: { value: SortBy; label: string }[] = [
    { value: "first_seen_at", label: "Newest" },
    { value: "price", label: "Price" },
    { value: "commute_time", label: "Commute" },
];

const currentSortLabel = computed(
    () => sortOptions.find((o) => o.value === sortBy.value)?.label ?? "Sort",
);

function setSort(col: SortBy, dir: SortDir) {
    if (sortBy.value === col && sortDir.value === dir) return;
    sortBy.value = col;
    sortDir.value = dir;
    loadInitial();
}

function toggleSortDir() {
    sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
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
    if (seconds == null) return "rgba(255,255,255,0.35)";
    const minutes = seconds / 60;
    if (minutes < 30) return "#2e7d32";
    if (minutes <= 60) return "#f9a825";
    return "#c62828";
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

onMounted(() => {
    teleportReady.value = true;
    loadInitial();
});

onBeforeUnmount(() => {
    panelObserver?.disconnect();
    mobileObserver?.disconnect();
});
</script>

<template>
    <Teleport :to="filterTarget" :disabled="!teleportReady">
        <v-menu :close-on-content-click="false" location="bottom start" offset="6">
            <template #activator="{ props }">
                <v-btn v-bind="props" rounded="pill" variant="outlined"
                    :color="activeFilterCount > 0 ? 'secondary' : undefined"
                    prepend-icon="mdi-filter-variant" class="filter-btn text-none"
                    :class="{ 'filter-btn--active': activeFilterCount > 0 }" size="small">
                    Filters
                    <v-badge v-if="activeFilterCount > 0" inline color="secondary" :content="activeFilterCount"
                        class="filter-btn__badge" />
                </v-btn>
            </template>
            <v-card min-width="260" class="filter-menu">
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Max price</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="maxPrice == null ? 'flat' : 'outlined'"
                            :color="maxPrice == null ? 'secondary' : undefined" @click="setMaxPrice(null)">No
                            max</v-chip>
                        <v-chip v-for="p in priceOptions" :key="p" size="small"
                            :variant="maxPrice === p ? 'flat' : 'outlined'"
                            :color="maxPrice === p ? 'secondary' : undefined"
                            @click="setMaxPrice(maxPrice === p ? null : p)">{{
                            formatCompactPrice(p) }}</v-chip>
                    </div>
                </div>
                <v-divider />
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Max commute</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="maxCommuteSec == null ? 'flat' : 'outlined'"
                            :color="maxCommuteSec == null ? 'secondary' : undefined" @click="setMaxCommute(null)">No
                            max</v-chip>
                        <v-chip v-for="m in commuteOptions" :key="m" size="small"
                            :variant="maxCommuteSec === m * 60 ? 'flat' : 'outlined'"
                            :color="maxCommuteSec === m * 60 ? 'secondary' : undefined"
                            @click="setMaxCommute(maxCommuteSec === m * 60 ? null : m)">{{ m }} min</v-chip>
                    </div>
                </div>
                <v-divider />
                <div class="filter-menu__section">
                    <div class="filter-menu__title">Recency</div>
                    <div class="filter-menu__chips">
                        <v-chip size="small" :variant="newWithinDays != null ? 'flat' : 'outlined'"
                            :color="newWithinDays != null ? 'secondary' : undefined" @click="toggleNewOnly">New today
                            only</v-chip>
                    </div>
                </div>
                <template v-if="activeFilterCount > 0">
                    <v-divider />
                    <div class="filter-menu__footer">
                        <v-btn variant="text" size="small" class="text-none" @click="clearAllFilters">Clear all</v-btn>
                    </div>
                </template>
            </v-card>
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
                :new-within-days="newWithinDays" @update:count="mapCount = $event" />
            <aside v-if="mdAndUp && drawerOpen" class="listings-side-panel">
                <div class="list-toolbar">
                    <div class="list-toolbar__count">
                        <span class="list-toolbar__count-num">{{ total.toLocaleString() }}</span>
                        <span class="list-toolbar__count-label">listing{{ total === 1 ? "" : "s" }}</span>
                    </div>
                    <div class="list-toolbar__actions">
                        <div id="drawer-filters-slot" class="list-toolbar__filters" />
                        <v-menu location="bottom end" offset="6">
                            <template #activator="{ props }">
                                <v-btn v-bind="props" rounded="pill" variant="outlined"
                                    class="list-toolbar__sort text-none" size="small"
                                    append-icon="mdi-menu-down">
                                    {{ currentSortLabel }}
                                </v-btn>
                            </template>
                            <v-list density="compact" min-width="180">
                                <v-list-item v-for="opt in sortOptions" :key="opt.value"
                                    :active="sortBy === opt.value" :title="opt.label"
                                    @click="setSort(opt.value, 'desc')" />
                                <v-divider />
                                <v-list-item :title="sortDir === 'asc' ? 'Descending' : 'Ascending'"
                                    :prepend-icon="sortDir === 'asc' ? 'mdi-arrow-down' : 'mdi-arrow-up'"
                                    @click="toggleSortDir" />
                            </v-list>
                        </v-menu>
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
            <div class="list-toolbar__count">
                <span class="list-toolbar__count-num">{{ total.toLocaleString() }}</span>
                <span class="list-toolbar__count-label">listing{{ total === 1 ? "" : "s" }}</span>
            </div>
            <div class="list-toolbar__actions">
                <div id="mobile-filters-slot" class="list-toolbar__filters" />
                <v-menu location="bottom end" offset="6">
                    <template #activator="{ props }">
                        <v-btn v-bind="props" rounded="pill" variant="outlined"
                            class="list-toolbar__sort text-none" size="small"
                            append-icon="mdi-menu-down">
                            {{ currentSortLabel }}
                        </v-btn>
                    </template>
                    <v-list density="compact" min-width="180">
                        <v-list-item v-for="opt in sortOptions" :key="opt.value"
                            :active="sortBy === opt.value" :title="opt.label"
                            @click="setSort(opt.value, 'desc')" />
                        <v-divider />
                        <v-list-item :title="sortDir === 'asc' ? 'Descending' : 'Ascending'"
                            :prepend-icon="sortDir === 'asc' ? 'mdi-arrow-down' : 'mdi-arrow-up'"
                            @click="toggleSortDir" />
                    </v-list>
                </v-menu>
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
                            <span class="listing-card__commute-dot"
                                :style="{ background: commuteColor(null) }" />
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

    <div v-if="!mdAndUp" class="mobile-bottom-card"
        :class="{ 'mobile-bottom-card--bare': viewMode === 'list' }">
        <div v-if="viewMode === 'map'" class="mobile-bottom-card__legend" aria-label="Transit commute time legend">
            <div class="mobile-bottom-card__legend-title">Commute to Downtown</div>
            <div class="mobile-bottom-card__legend-rows">
                <span class="mobile-bottom-card__legend-row">
                    <span class="mobile-bottom-card__swatch" style="background:#2e7d32" />
                    &lt; 30 min
                </span>
                <span class="mobile-bottom-card__legend-row">
                    <span class="mobile-bottom-card__swatch" style="background:#f9a825" />
                    30–60 min
                </span>
                <span class="mobile-bottom-card__legend-row">
                    <span class="mobile-bottom-card__swatch" style="background:#c62828" />
                    &gt; 60 min
                </span>
            </div>
        </div>
        <v-btn class="mobile-bottom-card__btn text-none" variant="tonal" rounded="lg" size="large" block
            @click="toggleViewMode">
            <v-icon start>{{ viewMode === "list" ? "mdi-map" : "mdi-format-list-bulleted" }}</v-icon>
            {{ viewMode === "list" ? "Show map" : "Show list" }}
        </v-btn>
    </div>
</template>

<style scoped>
.filter-btn {
    letter-spacing: normal;
    font-weight: 500;
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.85);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
    padding-inline: 12px;
    height: 30px;
}

.filter-btn--active {
    border-color: rgba(var(--v-theme-secondary), 0.55);
}

.filter-btn :deep(.v-btn__prepend) {
    margin-inline-end: 6px;
}

.filter-btn__badge {
    margin-inline-start: 8px;
}

.filter-menu {
    padding: 4px 0;
}

.filter-menu__section {
    padding: 12px 14px;
}

.filter-menu__title {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(var(--v-theme-on-surface), 0.6);
    margin-bottom: 8px;
}

.filter-menu__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
}

.filter-menu__footer {
    display: flex;
    justify-content: flex-end;
    padding: 6px 8px;
}

.listings-mobile {
    padding-bottom: calc(96px + env(safe-area-inset-bottom, 0px));
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
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 12px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    flex: 0 0 auto;
}

.list-toolbar--mobile {
    padding: 12px 14px 10px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.list-toolbar__count {
    display: flex;
    flex-direction: column;
    line-height: 1;
}

.list-toolbar__count-num {
    font-size: 0.9375rem;
    font-weight: 700;
    letter-spacing: -0.01em;
}

.list-toolbar__count-label {
    font-size: 0.625rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin-top: 2px;
}

.list-toolbar__actions {
    display: flex;
    align-items: center;
    gap: 12px;
}

.list-toolbar__filters:empty {
    display: none;
}

.list-toolbar__sort {
    letter-spacing: normal;
    font-weight: 500;
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.85);
    border-color: rgba(var(--v-theme-on-surface), 0.18);
    padding-inline: 12px;
    height: 30px;
}

.list-toolbar__sort :deep(.v-btn__append) {
    margin-inline-start: 4px;
    opacity: 0.7;
}

.mobile-bottom-card {
    position: fixed;
    left: 12px;
    right: 12px;
    bottom: calc(28px + env(safe-area-inset-bottom, 0px));
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px;
    border-radius: 14px;
    background-color: rgba(20, 22, 28, 0.78);
    border: 1px solid rgba(255, 255, 255, 0.08);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.45);
}

.mobile-bottom-card--bare {
    padding: 12px;
    background-color: rgba(20, 22, 28, 0.6);
    border-color: rgba(255, 255, 255, 0.06);
}

.mobile-bottom-card__legend {
    display: flex;
    flex-direction: column;
    gap: 6px;
    color: rgba(255, 255, 255, 0.92);
}

.mobile-bottom-card__legend-title {
    font-size: 0.875rem;
    font-weight: 600;
}

.mobile-bottom-card__legend-rows {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.8);
}

.mobile-bottom-card__legend-row {
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.mobile-bottom-card__swatch {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.3);
}

.mobile-bottom-card__btn {
    letter-spacing: normal;
    font-weight: 600;
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
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.3);
}

.listing-card__seen {
    margin-left: auto;
    color: rgba(var(--v-theme-on-surface), 0.45);
    font-size: 0.75rem;
}
</style>
