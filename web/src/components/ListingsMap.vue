<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef, watch } from "vue";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import "leaflet.markercluster";
import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";
import { latLngToCell, cellToBoundary } from "h3-js";
import { listListingsForMap, type ListingMapPin } from "../api/listings";
import { listTransitStops } from "../api/transit";

const props = defineProps<{
    maxPrice: number | null;
    maxCommuteSec: number | null;
    newWithinDays: number | null;
}>();

const emit = defineEmits<{
    (e: "update:count", n: number): void;
    (e: "pin-click", payload: { board: number; mls: number }): void;
}>();

const mapEl = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const cluster = shallowRef<L.MarkerClusterGroup | null>(null);
const hexLayer = shallowRef<L.LayerGroup | null>(null);
const HEX_RESOLUTION = 8;
const loading = ref(false);
const error = ref<string | null>(null);
let hasFitBounds = false;
let resizeObserver: ResizeObserver | null = null;
const markersByKey = new Map<string, L.Marker>();

function listingKey(board: number, mls: number): string {
    return `${board}-${mls}`;
}

const MONTREAL_CENTER: L.LatLngTuple = [45.5048, -73.5772];
const MCGILL_STATION: L.LatLngTuple = [45.5045, -73.5746];

function formatCompactPrice(price: number | null): string {
    if (price == null) return "—";
    if (price >= 1_000_000) {
        const m = price / 1_000_000;
        return `$${m % 1 === 0 ? m.toFixed(0) : m.toFixed(1)}M`;
    }
    return `$${Math.round(price / 1000)}k`;
}

function commuteTier(seconds: number | null): "fast" | "mid" | "slow" | "unknown" {
    if (seconds == null) return "unknown";
    const minutes = seconds / 60;
    if (minutes < 30) return "fast";
    if (minutes <= 60) return "mid";
    return "slow";
}

type MarkerData = { price: number | null; commuteSec: number | null };

function median(nums: number[]): number | null {
    if (nums.length === 0) return null;
    const sorted = [...nums].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 === 0
        ? (sorted[mid - 1] + sorted[mid]) / 2
        : sorted[mid];
}

function clusterIcon(c: L.MarkerCluster): L.DivIcon {
    const commutes: number[] = [];
    for (const m of c.getAllChildMarkers()) {
        const d = (m as L.Marker & { _data?: MarkerData })._data;
        if (d?.commuteSec != null) commutes.push(d.commuteSec);
    }
    const medCommute = median(commutes);
    const tier = commuteTier(medCommute);
    const count = c.getChildCount();
    const sizeClass = count >= 100 ? "lg" : count >= 10 ? "md" : "sm";
    return L.divIcon({
        className: `price-cluster price-cluster--${tier} price-cluster--${sizeClass}`,
        html: `<div class="price-cluster__inner"><span class="price-cluster__count">${count}</span></div>`,
        iconSize: L.point(40, 40),
    });
}

function pricePillIcon(
    price: number | null,
    commuteSec: number | null,
): L.DivIcon {
    const label = formatCompactPrice(price);
    const tier = commuteTier(commuteSec);
    return L.divIcon({
        className: `price-pin price-pin--${tier}`,
        html: `<span class="price-pin__label">${label}</span>`,
        iconSize: undefined as unknown as L.PointExpression,
        iconAnchor: [0, 0],
    });
}

function formatPrice(price: number | null): string {
    if (price == null) return "—";
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
        maximumFractionDigits: 0,
    }).format(price);
}

function escapeHtml(s: string): string {
    return s
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

function parseAddress(raw: string | null | undefined): {
    street: string;
    locality: string;
} {
    if (!raw) return { street: "—", locality: "" };
    const parts = raw.split("|").map((s) => s.trim()).filter(Boolean);
    const street = parts[0] ?? raw;
    const locality = parts.slice(1).join(", ");
    return { street, locality };
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

function popupHtml(pin: ListingMapPin): string {
    const { street, locality } = parseAddress(pin.address);
    const streetEsc = escapeHtml(street);
    const localityEsc = escapeHtml(locality);
    const price = escapeHtml(formatPrice(pin.currentPrice));
    const tier = commuteTier(pin.commuteSecondsDowntown);
    const commuteMin =
        pin.commuteSecondsDowntown != null
            ? Math.round(pin.commuteSecondsDowntown / 60).toString()
            : "—";
    const slug = escapeHtml(pin.slug);
    const directionsHref = commuteMapUrl(pin.address);
    const directionsBtn = directionsHref
        ? `<a href="${escapeHtml(directionsHref)}" target="_blank" rel="noopener noreferrer" class="map-popup__btn map-popup__btn--secondary">
                <i class="mdi mdi-directions" aria-hidden="true"></i><span>Directions</span>
            </a>`
        : "";
    const localityLine = localityEsc
        ? `<div class="map-popup__locality">${localityEsc}</div>`
        : "";
    return `
        <div class="map-popup">
            <div class="map-popup__chip map-popup__chip--${tier}">${price}</div>
            <div class="map-popup__address">
                <div class="map-popup__street">${streetEsc}</div>
                ${localityLine}
            </div>
            <div class="map-popup__commute map-popup__commute--${tier}">
                <i class="mdi mdi-subway-variant map-popup__commute-icon" aria-hidden="true"></i>
                <div class="map-popup__commute-text">
                    <div class="map-popup__commute-value">${commuteMin}<span class="map-popup__commute-unit"> min</span></div>
                    <div class="map-popup__commute-dest">to McGill Station</div>
                </div>
            </div>
            <div class="map-popup__actions">
                <a href="${slug}" target="_blank" rel="noopener noreferrer" class="map-popup__btn map-popup__btn--primary">
                    <i class="mdi mdi-open-in-new" aria-hidden="true"></i><span>Listing</span>
                </a>
                ${directionsBtn}
            </div>
        </div>
    `;
}

async function load() {
    if (!map.value || !cluster.value) return;
    loading.value = true;
    error.value = null;
    try {
        const pins = await listListingsForMap({
            ...(props.maxPrice != null && { maxPrice: props.maxPrice }),
            ...(props.maxCommuteSec != null && {
                maxCommuteSec: props.maxCommuteSec,
            }),
            ...(props.newWithinDays != null && {
                newWithinDays: props.newWithinDays,
            }),
        });
        clearHighlight();
        cluster.value.clearLayers();
        markersByKey.clear();
        const markers: L.Marker[] = [];
        for (const pin of pins) {
            const m = L.marker([pin.latitude, pin.longitude], {
                icon: pricePillIcon(pin.currentPrice, pin.commuteSecondsDowntown),
                riseOnHover: true,
            }) as L.Marker & { _data?: MarkerData };
            m._data = {
                price: pin.currentPrice,
                commuteSec: pin.commuteSecondsDowntown,
            };
            m.bindPopup(popupHtml(pin));
            m.on("click", () => emit("pin-click", { board: pin.board, mls: pin.mls }));
            markers.push(m);
            markersByKey.set(listingKey(pin.board, pin.mls), m);
        }
        cluster.value.addLayers(markers);
        emit("update:count", markers.length);
        if (markers.length > 0 && !hasFitBounds) {
            const bounds = L.latLngBounds(
                markers.map((m) => m.getLatLng()),
            );
            map.value.fitBounds(bounds, { padding: [40, 40], maxZoom: 14 });
            hasFitBounds = true;
        }
    } catch (e: any) {
        error.value =
            e?.response?.data?.error ??
            e?.message ??
            "failed to load map listings";
    } finally {
        loading.value = false;
    }
}

function gradientColor(commuteSec: number): string {
    const minutes = commuteSec / 60;
    const t = Math.max(0, Math.min(1, minutes / 90));
    const hue = 120 * (1 - t);
    return `hsl(${hue.toFixed(0)}, 70%, 45%)`;
}

function medianOf(nums: number[]): number {
    const sorted = [...nums].sort((a, b) => a - b);
    const mid = Math.floor(sorted.length / 2);
    return sorted.length % 2 === 0
        ? (sorted[mid - 1] + sorted[mid]) / 2
        : sorted[mid];
}

function buildHexLayer(stops: { latitude: number; longitude: number; commuteSec: number }[]): L.LayerGroup {
    const byCell = new Map<string, number[]>();
    for (const s of stops) {
        const cell = latLngToCell(s.latitude, s.longitude, HEX_RESOLUTION);
        const arr = byCell.get(cell);
        if (arr) arr.push(s.commuteSec);
        else byCell.set(cell, [s.commuteSec]);
    }
    const layer = L.layerGroup();
    for (const [cell, commutes] of byCell) {
        const med = medianOf(commutes);
        const boundary = cellToBoundary(cell, false) as [number, number][];
        const poly = L.polygon(boundary, {
            pane: "hexPane",
            color: gradientColor(med),
            weight: 0,
            fillColor: gradientColor(med),
            fillOpacity: 0.15,
            interactive: false,
        });
        layer.addLayer(poly);
    }
    return layer;
}

async function loadStops() {
    if (!map.value) return;
    try {
        const stops = await listTransitStops();
        hexLayer.value = buildHexLayer(stops);
        hexLayer.value.addTo(map.value);
    } catch (e) {
        console.warn("failed to load transit stops layer", e);
    }
}

onMounted(() => {
    if (!mapEl.value) return;
    map.value = L.map(mapEl.value, {
        center: MONTREAL_CENTER,
        zoom: 11,
        zoomControl: false,
    });
    map.value.createPane("hexPane");
    const hexPane = map.value.getPane("hexPane");
    if (hexPane) {
        hexPane.style.zIndex = "350";
        hexPane.style.pointerEvents = "none";
    }
    L.tileLayer(
        "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
        {
            attribution:
                '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/attributions">CARTO</a>',
            subdomains: "abcd",
            maxZoom: 19,
        },
    ).addTo(map.value);
    cluster.value = L.markerClusterGroup({
        showCoverageOnHover: false,
        spiderfyOnMaxZoom: true,
        maxClusterRadius: 50,
        disableClusteringAtZoom: 14,
        iconCreateFunction: clusterIcon,
    });
    loadStops();
    map.value.addLayer(cluster.value);
    L.marker(MCGILL_STATION, {
        icon: L.divIcon({
            className: "downtown-target",
            html: `<span class="downtown-target__star" aria-hidden="true">★</span>`,
            iconSize: [22, 22],
            iconAnchor: [11, 11],
        }),
        interactive: true,
        keyboard: false,
        zIndexOffset: -500,
    })
        .bindTooltip("McGill Station", {
            direction: "top",
            offset: [0, -8],
            className: "downtown-target__tooltip",
            opacity: 1,
        })
        .addTo(map.value);
    load();

    resizeObserver = new ResizeObserver(() => {
        map.value?.invalidateSize();
    });
    resizeObserver.observe(mapEl.value);
});

onBeforeUnmount(() => {
    if (resizeObserver) {
        resizeObserver.disconnect();
        resizeObserver = null;
    }
    if (map.value) {
        map.value.remove();
        map.value = null;
        cluster.value = null;
        hexLayer.value = null;
        markersByKey.clear();
    }
});

watch(
    () => [props.maxPrice, props.maxCommuteSec, props.newWithinDays],
    () => load(),
);

function focusListing(board: number, mls: number): boolean {
    const marker = markersByKey.get(listingKey(board, mls));
    if (!marker || !map.value || !cluster.value) return false;
    cluster.value.zoomToShowLayer(marker, () => {
        marker.openPopup();
    });
    return true;
}

let highlightedEl: HTMLElement | null = null;

function highlightListing(board: number, mls: number): void {
    clearHighlight();
    const marker = markersByKey.get(listingKey(board, mls));
    if (!marker || !cluster.value) return;
    const visible = cluster.value.getVisibleParent(marker) ?? marker;
    const el = (visible as L.Marker | L.MarkerCluster).getElement?.();
    if (!el) return;
    el.classList.add("price-pin--highlighted");
    highlightedEl = el;
}

function clearHighlight(): void {
    if (highlightedEl) {
        highlightedEl.classList.remove("price-pin--highlighted");
        highlightedEl = null;
    }
}

defineExpose({ focusListing, highlightListing, clearHighlight });
</script>

<template>
    <div class="listings-map-wrap">
        <div v-if="loading || error" class="listings-map-status">
            <v-progress-circular v-if="loading" indeterminate size="20" width="2" class="mr-2" />
            <v-alert v-if="error" type="error" density="compact" variant="tonal" class="ma-0">{{ error }}</v-alert>
        </div>
        <div ref="mapEl" class="listings-map" />
        <div class="listings-map-legend" aria-label="Transit commute time legend">
            <div class="listings-map-legend__title">Commute to downtown</div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch" style="background:#2e7d32" />
                &lt; 30 min
            </div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch" style="background:#f9a825" />
                30–60 min
            </div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch" style="background:#c62828" />
                &gt; 60 min
            </div>
        </div>
    </div>
</template>

<style scoped>
.listings-map-wrap {
    position: relative;
    display: flex;
    flex: 1 1 auto;
    min-height: 0;
    width: 100%;
}

.listings-map {
    width: 100%;
    flex: 1 1 auto;
    min-height: 320px;
    border-radius: 4px;
    overflow: hidden;
}

.listings-map-status {
    position: absolute;
    top: 12px;
    right: 12px;
    z-index: 500;
    display: flex;
    align-items: center;
    background-color: rgba(var(--v-theme-surface), 0.85);
    padding: 6px 12px;
    border-radius: 999px;
    backdrop-filter: blur(4px);
    pointer-events: none;
}

.listings-map-legend {
    position: absolute;
    bottom: 24px;
    right: 12px;
    z-index: 500;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    background-color: rgba(20, 20, 24, 0.6);
    color: rgba(255, 255, 255, 0.92);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    font-size: 11px;
    line-height: 1.2;
    pointer-events: none;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);
}

.listings-map-legend__title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(255, 255, 255, 0.65);
    margin-bottom: 2px;
}

.listings-map-legend__row {
    display: flex;
    align-items: center;
    gap: 6px;
}

.listings-map-legend__swatch {
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 50%;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.3);
}

@media (max-width: 959.98px) {
    .listings-map-legend {
        display: none;
    }
}
</style>

<style>
.price-pin {
    background: transparent;
    border: 0;
    --pin-ring: rgba(255, 255, 255, 0.18);
}

.downtown-target {
    background: transparent;
    border: 0;
    pointer-events: auto;
}

.downtown-target__star {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    font-size: 18px;
    line-height: 1;
    color: rgba(255, 255, 255, 0.85);
    text-shadow:
        0 0 4px rgba(0, 0, 0, 0.85),
        0 0 8px rgba(0, 0, 0, 0.6),
        0 1px 1px rgba(0, 0, 0, 0.9);
    cursor: help;
    transition: transform 120ms ease, color 120ms ease;
}

.downtown-target:hover .downtown-target__star {
    color: #ffd54f;
    transform: scale(1.15);
}

.leaflet-tooltip.downtown-target__tooltip {
    background: rgba(20, 22, 28, 0.82);
    color: rgba(255, 255, 255, 0.92);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.02em;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.35);
    white-space: nowrap;
}

.leaflet-tooltip.downtown-target__tooltip::before {
    border-top-color: rgba(20, 22, 28, 0.82);
}

.price-pin--fast {
    --pin-ring: rgba(76, 175, 80, 0.9);
}

.price-pin--mid {
    --pin-ring: rgba(255, 179, 0, 0.9);
}

.price-pin--slow {
    --pin-ring: rgba(239, 83, 80, 0.9);
}

.price-pin--unknown {
    --pin-ring: rgba(255, 255, 255, 0.22);
}

.price-pin__label {
    position: relative;
    display: inline-block;
    transform: translate(-50%, -50%);
    padding: 4px 11px;
    border-radius: 999px;
    background: linear-gradient(180deg, rgba(34, 36, 44, 0.88) 0%, rgba(18, 20, 26, 0.88) 100%);
    color: #fff;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.01em;
    line-height: 1.15;
    white-space: nowrap;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow:
        0 4px 10px rgba(0, 0, 0, 0.45),
        0 1px 2px rgba(0, 0, 0, 0.35),
        inset 0 0 0 1px rgba(255, 255, 255, 0.06),
        inset 0 1px 0 rgba(255, 255, 255, 0.08);
    cursor: pointer;
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.price-pin__label::before {
    content: "";
    position: absolute;
    inset: -2px;
    border-radius: 999px;
    border: 1.5px solid var(--pin-ring);
    pointer-events: none;
}

.price-pin:hover .price-pin__label {
    transform: translate(-50%, -50%) scale(1.08);
    box-shadow:
        0 6px 14px rgba(0, 0, 0, 0.55),
        0 1px 2px rgba(0, 0, 0, 0.4),
        inset 0 0 0 1px rgba(255, 255, 255, 0.1);
    z-index: 1000;
}

.price-pin.price-pin--highlighted {
    z-index: 1001 !important;
}

.price-pin.price-pin--highlighted .price-pin__label {
    transform: translate(-50%, -50%) scale(1.18);
}

.price-pin.price-pin--highlighted .price-pin__label::before {
    border-color: rgb(var(--v-theme-secondary));
    border-width: 2px;
    inset: -3px;
}

.price-cluster {
    background: transparent;
    border: 0;
    --cluster-ring: rgba(255, 255, 255, 0.18);
}

.price-cluster--fast {
    --cluster-ring: rgba(76, 175, 80, 0.85);
}

.price-cluster--mid {
    --cluster-ring: rgba(255, 179, 0, 0.85);
}

.price-cluster--slow {
    --cluster-ring: rgba(239, 83, 80, 0.85);
}

.price-cluster--unknown {
    --cluster-ring: rgba(255, 255, 255, 0.25);
}

.price-cluster__inner {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    background-color: rgba(20, 22, 28, 0.82);
    color: #fff;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow:
        0 4px 10px rgba(0, 0, 0, 0.45),
        0 1px 2px rgba(0, 0, 0, 0.35),
        inset 0 0 0 1px rgba(255, 255, 255, 0.06);
    cursor: pointer;
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.price-cluster__inner::before {
    content: "";
    position: absolute;
    inset: -3px;
    border-radius: 50%;
    border: 2px solid var(--cluster-ring);
    pointer-events: none;
}

.price-cluster--md .price-cluster__inner {
    width: 42px;
    height: 42px;
}

.price-cluster--lg .price-cluster__inner {
    width: 52px;
    height: 52px;
}

.price-cluster:hover .price-cluster__inner {
    transform: scale(1.06);
    box-shadow:
        0 6px 14px rgba(0, 0, 0, 0.55),
        0 1px 2px rgba(0, 0, 0, 0.4),
        inset 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.price-cluster__count {
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.01em;
    line-height: 1;
    color: rgba(255, 255, 255, 0.96);
}

.price-cluster--md .price-cluster__count {
    font-size: 13px;
}

.price-cluster--lg .price-cluster__count {
    font-size: 15px;
}

.marker-cluster.price-pin--highlighted {
    z-index: 1001 !important;
}

.marker-cluster.price-pin--highlighted>div,
.price-cluster.price-pin--highlighted .price-cluster__inner {
    box-shadow:
        0 0 0 3px rgb(var(--v-theme-secondary)),
        0 6px 14px rgba(0, 0, 0, 0.6);
    transform: scale(1.12);
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.leaflet-popup-content-wrapper {
    background-color: rgb(var(--v-theme-surface));
    color: rgb(var(--v-theme-on-surface));
    border-radius: 12px;
    box-shadow:
        0 10px 28px rgba(0, 0, 0, 0.45),
        0 0 0 1px rgba(255, 255, 255, 0.06);
    padding: 2px;
}

.leaflet-popup-content {
    margin: 14px 16px;
}

.leaflet-popup-tip {
    background-color: rgb(var(--v-theme-surface));
}

.leaflet-container a.leaflet-popup-close-button {
    color: rgba(var(--v-theme-on-surface), 0.65);
    padding: 6px 8px 0 0;
}

.map-popup {
    min-width: 260px;
    max-width: 300px;
    font-size: 0.875rem;
}

.map-popup__chip {
    display: inline-block;
    padding: 4px 11px;
    border-radius: 999px;
    font-size: 0.95rem;
    font-weight: 700;
    line-height: 1.2;
    margin-bottom: 12px;
    background-color: var(--chip-bg);
    color: var(--chip-fg);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.35);
}

.map-popup__chip--fast {
    --chip-bg: #2e7d32;
    --chip-fg: #ffffff;
}

.map-popup__chip--mid {
    --chip-bg: #f9a825;
    --chip-fg: #1a1a1a;
}

.map-popup__chip--slow {
    --chip-bg: #c62828;
    --chip-fg: #ffffff;
}

.map-popup__chip--unknown {
    --chip-bg: #555555;
    --chip-fg: #ffffff;
}

.map-popup__address {
    margin-bottom: 12px;
}

.map-popup__street {
    font-size: 0.95rem;
    font-weight: 600;
    line-height: 1.35;
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.map-popup__locality {
    font-size: 0.78rem;
    line-height: 1.3;
    color: rgba(var(--v-theme-on-surface), 0.6);
    margin-top: 2px;
}

.map-popup__commute {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    margin-bottom: 12px;
    border-radius: 8px;
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    --commute-accent: rgba(var(--v-theme-on-surface), 0.7);
}

.map-popup__commute--fast {
    --commute-accent: #4caf50;
}

.map-popup__commute--mid {
    --commute-accent: #ffb300;
}

.map-popup__commute--slow {
    --commute-accent: #ef5350;
}

.map-popup__commute-icon {
    font-size: 22px;
    color: var(--commute-accent);
    flex: 0 0 auto;
}

.map-popup__commute-text {
    min-width: 0;
}

.map-popup__commute-value {
    font-size: 1.05rem;
    font-weight: 700;
    line-height: 1.1;
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.map-popup__commute-unit {
    font-size: 0.8rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.7);
}

.map-popup__commute-dest {
    margin-top: 2px;
    font-size: 0.66rem;
    letter-spacing: 0.07em;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.55);
}

.map-popup__actions {
    display: flex;
    gap: 6px;
}

.map-popup__btn {
    flex: 1 1 0;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    padding: 8px 10px;
    border-radius: 8px;
    font-size: 0.8rem;
    font-weight: 600;
    text-decoration: none;
    white-space: nowrap;
    transition:
        filter 120ms ease,
        transform 120ms ease,
        background-color 120ms ease;
}

.map-popup__btn .mdi {
    font-size: 16px;
}

.map-popup__btn--primary {
    background-color: rgb(var(--v-theme-secondary));
    color: rgb(var(--v-theme-on-secondary));
}

.map-popup__btn--primary:hover {
    filter: brightness(1.1);
}

.map-popup__btn--secondary {
    background-color: transparent;
    color: rgba(var(--v-theme-on-surface), 0.9);
    box-shadow: inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.22);
}

.map-popup__btn--secondary:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.06);
}

.map-popup__btn:active {
    transform: translateY(1px);
}
</style>
