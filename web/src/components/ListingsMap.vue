<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, shallowRef, watch } from "vue";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import "leaflet.markercluster";
import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";
import { listListingsForMap, type ListingMapPin } from "../api/listings";
import { listTransitStops } from "../api/transit";

const props = defineProps<{
    maxPrice: number | null;
    maxCommuteSec: number | null;
    newWithinDays: number | null;
}>();

const mapEl = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const cluster = shallowRef<L.MarkerClusterGroup | null>(null);
const stopsLayer = shallowRef<L.LayerGroup | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);
const pinCount = ref(0);

const MONTREAL_CENTER: L.LatLngTuple = [45.5048, -73.5772];

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

function formatCommute(seconds: number | null): string {
    if (seconds == null) return "—";
    return `${Math.round(seconds / 60)} min`;
}

function escapeHtml(s: string): string {
    return s
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

function popupHtml(pin: ListingMapPin): string {
    const address = escapeHtml(pin.address ?? "—");
    const price = escapeHtml(formatPrice(pin.currentPrice));
    const commute = escapeHtml(formatCommute(pin.commuteSecondsDowntown));
    const slug = escapeHtml(pin.slug);
    return `
        <div class="map-popup">
            <div class="map-popup__price">${price}</div>
            <div class="map-popup__address">${address}</div>
            <div class="map-popup__commute">
                <span class="map-popup__label">Commute:</span> ${commute}
            </div>
            <a href="${slug}" target="_blank" rel="noopener noreferrer" class="map-popup__link">
                View on Realtor.ca →
            </a>
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
        cluster.value.clearLayers();
        const markers: L.Marker[] = [];
        for (const pin of pins) {
            const m = L.marker([pin.latitude, pin.longitude], {
                icon: pricePillIcon(pin.currentPrice, pin.commuteSecondsDowntown),
                riseOnHover: true,
            });
            m.bindPopup(popupHtml(pin));
            markers.push(m);
        }
        cluster.value.addLayers(markers);
        pinCount.value = markers.length;
        if (markers.length > 0) {
            const bounds = L.latLngBounds(
                markers.map((m) => m.getLatLng()),
            );
            map.value.fitBounds(bounds, { padding: [40, 40], maxZoom: 14 });
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

function stopColor(commuteSec: number): string {
    const minutes = commuteSec / 60;
    if (minutes < 30) return "#2e7d32";
    if (minutes <= 60) return "#f9a825";
    return "#c62828";
}

function radiusForZoom(zoom: number): number {
    if (zoom <= 9) return 1.5;
    if (zoom >= 15) return 7;
    return 1.5 + (zoom - 9) * 0.9;
}

const stopCircles: L.CircleMarker[] = [];

function applyStopRadius() {
    if (!map.value) return;
    const r = radiusForZoom(map.value.getZoom());
    for (const c of stopCircles) c.setRadius(r);
}

async function loadStops() {
    if (!map.value) return;
    try {
        const stops = await listTransitStops();
        const renderer = L.canvas({ padding: 0.5 });
        const layer = L.layerGroup();
        const r = radiusForZoom(map.value.getZoom());
        stopCircles.length = 0;
        for (const s of stops) {
            const circle = L.circleMarker([s.latitude, s.longitude], {
                renderer,
                radius: r,
                stroke: false,
                fillColor: stopColor(s.commuteSec),
                fillOpacity: 0.6,
                interactive: false,
            });
            layer.addLayer(circle);
            stopCircles.push(circle);
        }
        stopsLayer.value = layer;
        layer.addTo(map.value);
        map.value.on("zoomend", applyStopRadius);
    } catch (e) {
        console.warn("failed to load transit stops layer", e);
    }
}

onMounted(() => {
    if (!mapEl.value) return;
    map.value = L.map(mapEl.value, {
        center: MONTREAL_CENTER,
        zoom: 11,
    });
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
    });
    loadStops();
    map.value.addLayer(cluster.value);
    load();
});

onBeforeUnmount(() => {
    if (map.value) {
        map.value.off("zoomend", applyStopRadius);
        map.value.remove();
        map.value = null;
        cluster.value = null;
        stopsLayer.value = null;
        stopCircles.length = 0;
    }
});

watch(
    () => [props.maxPrice, props.maxCommuteSec, props.newWithinDays],
    () => load(),
);
</script>

<template>
    <div class="listings-map-wrap">
        <div class="listings-map-status">
            <v-progress-circular v-if="loading" indeterminate size="20" width="2" class="mr-2" />
            <span v-if="!loading && !error" class="text-body-2 text-medium-emphasis">
                {{ pinCount }} on map
            </span>
            <v-alert v-if="error" type="error" density="compact" variant="tonal" class="ma-0">{{ error }}</v-alert>
        </div>
        <div ref="mapEl" class="listings-map" />
    </div>
</template>

<style scoped>
.listings-map-wrap {
    position: relative;
}

.listings-map {
    width: 100%;
    height: calc(100vh - 280px);
    min-height: 480px;
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
</style>

<style>
.price-pin {
    background: transparent;
    border: 0;
    --pin-bg: rgb(var(--v-theme-secondary));
    --pin-fg: rgb(var(--v-theme-on-secondary));
}

.price-pin--fast {
    --pin-bg: #2e7d32;
    --pin-fg: #ffffff;
}

.price-pin--mid {
    --pin-bg: #f9a825;
    --pin-fg: #1a1a1a;
}

.price-pin--slow {
    --pin-bg: #c62828;
    --pin-fg: #ffffff;
}

.price-pin--unknown {
    --pin-bg: #555;
    --pin-fg: #ffffff;
}

.price-pin__label {
    display: inline-block;
    transform: translate(-50%, -100%);
    padding: 4px 9px;
    border-radius: 999px;
    background-color: var(--pin-bg);
    color: var(--pin-fg);
    font-size: 12px;
    font-weight: 600;
    line-height: 1.1;
    white-space: nowrap;
    box-shadow:
        0 1px 2px rgba(0, 0, 0, 0.45),
        0 0 0 1px rgba(0, 0, 0, 0.25);
    cursor: pointer;
    transition: transform 100ms ease, box-shadow 100ms ease;
}

.price-pin__label::after {
    content: "";
    position: absolute;
    left: 50%;
    bottom: -4px;
    transform: translateX(-50%);
    width: 0;
    height: 0;
    border-left: 4px solid transparent;
    border-right: 4px solid transparent;
    border-top: 4px solid var(--pin-bg);
}

.price-pin:hover .price-pin__label {
    transform: translate(-50%, -100%) scale(1.06);
    box-shadow:
        0 2px 6px rgba(0, 0, 0, 0.55),
        0 0 0 1px rgba(0, 0, 0, 0.35);
    z-index: 1000;
}

.leaflet-popup-content-wrapper {
    background-color: rgb(var(--v-theme-surface));
    color: rgb(var(--v-theme-on-surface));
    border-radius: 8px;
}

.leaflet-popup-tip {
    background-color: rgb(var(--v-theme-surface));
}

.leaflet-container a.leaflet-popup-close-button {
    color: rgba(var(--v-theme-on-surface), 0.65);
}

.map-popup {
    min-width: 180px;
    font-size: 0.875rem;
}

.map-popup__price {
    font-size: 1.05rem;
    font-weight: 600;
    margin-bottom: 4px;
}

.map-popup__address {
    color: rgba(var(--v-theme-on-surface), 0.85);
    margin-bottom: 8px;
    line-height: 1.35;
}

.map-popup__commute {
    color: rgba(var(--v-theme-on-surface), 0.75);
    margin-bottom: 10px;
}

.map-popup__label {
    color: rgba(var(--v-theme-on-surface), 0.55);
}

.map-popup__link {
    color: rgb(var(--v-theme-secondary));
    text-decoration: none;
    font-weight: 500;
}

.map-popup__link:hover {
    text-decoration: underline;
}
</style>
