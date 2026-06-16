<script setup lang="ts">
import { computed, nextTick, onMounted, onBeforeUnmount, ref, shallowRef, watch } from "vue";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import "leaflet.markercluster";
import "leaflet.markercluster/dist/MarkerCluster.css";
import "leaflet.markercluster/dist/MarkerCluster.Default.css";
import { latLngToCell, cellToBoundary } from "h3-js";
import { useDisplay } from "vuetify";
import { listListingsForMap, type ListingMapPin } from "../api/listings";
import { listTransitStops } from "../api/transit";
import { useBottomSheetDrag } from "../composables/useBottomSheetDrag";
import { useMobileBottomSheetCoordinator } from "../composables/useMobileBottomSheetCoordinator";
import { debounce } from "../utils/debounce";

const props = defineProps<{
    buildingTypes: number[];
    maxPrice: number | null;
    maxCommuteSec: number | null;
    newWithinDays: number | null;
    minBedrooms: number | null;
    minBathrooms: number | null;
    minInteriorAreaSqft: number | null;
    favoritesOnly: boolean;
    includeExpired: boolean;
    mobileSheet: boolean;
}>();

const emit = defineEmits<{
    (e: "update:count", n: number): void;
    (e: "update:loading", v: boolean): void;
    (e: "pin-click", payload: { board: number; mls: number }): void;
    (e: "toggle-favorite", payload: { board: number; mls: number; isFavorite: boolean }): void;
    (e: "error", message: string): void;
}>();

let forceRefit = false;

const { mobile } = useDisplay();

const mapEl = ref<HTMLElement | null>(null);
const map = shallowRef<L.Map | null>(null);
const cluster = shallowRef<L.MarkerClusterGroup | null>(null);
const hexLayer = shallowRef<L.LayerGroup | null>(null);
const HEX_RESOLUTION = 8;
const loading = ref(false);
const visiblePins = ref<ListingMapPin[]>([]);
const selectedPin = ref<ListingMapPin | null>(null);
const animateMobileSheet = ref(true);
const {
    dragClasses: sheetDragClasses,
    dragStyle: sheetDragStyle,
    resetDrag: resetSheetDrag,
    onDragPointerDown,
    onDragPointerMove,
    onDragPointerUp,
    onDragPointerCancel,
} = useBottomSheetDrag(() => closeSheet({ immediate: true }));
useMobileBottomSheetCoordinator({
    open: () => selectedPin.value != null,
    close: () => closeSheet({ immediate: true }),
    enabled: () => props.mobileSheet,
});
const swipeStageEl = ref<HTMLElement | null>(null);
const swipeOffsetX = ref(0);
const swipeTransition = ref(false);
const swipeState = ref<"idle" | "pending" | "horizontal" | "vertical">("idle");
const failedPhotoKeys = ref(new Set<string>());
let swipePointerId: number | null = null;
let swipeCaptureEl: HTMLElement | null = null;
let swipeStartX = 0;
let swipeStartY = 0;
let swipeLastX = 0;
let swipeLastAt = 0;
let swipeVelocityX = 0;
let swipeTimer: number | null = null;
let hasFitBounds = false;
let resizeObserver: ResizeObserver | null = null;
const markersByKey = new Map<string, L.Marker>();
let selectedMarkerEl: HTMLElement | null = null;

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

type MarkerData = {
    board: number;
    mls: number;
    price: number | null;
    commuteSec: number | null;
    isFavorite: boolean;
};

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
    const mobileClass = mobile.value ? " price-cluster--mobile" : "";
    const iconSize =
        sizeClass === "lg"
            ? L.point(66, mobile.value ? 40 : 38)
            : sizeClass === "md"
                ? L.point(58, mobile.value ? 40 : 36)
                : L.point(mobile.value ? 60 : 52, mobile.value ? 38 : 34);
    const listingLabel = count === 1 ? "listing" : "listings";
    const html = mobile.value
        ? `<div class="price-cluster__inner"><span class="price-cluster__count">${count}</span><span class="price-cluster__label">${listingLabel}</span></div>`
        : `<div class="price-cluster__inner"><span class="price-cluster__count">${count}</span><span class="price-cluster__label">listings</span></div>`;
    return L.divIcon({
        className: `price-cluster price-cluster--${tier} price-cluster--${sizeClass}${mobileClass}`,
        html,
        iconSize,
    });
}

function pricePillIcon(
    price: number | null,
    commuteSec: number | null,
    buildingType: number,
): L.DivIcon {
    const label = formatCompactPrice(price);
    const tier = commuteTier(commuteSec);
    const icon = buildingType === 1 ? "mdi-home" : "mdi-domain";
    return L.divIcon({
        className: `price-pin price-pin--${tier}`,
        html: `<span class="price-pin__label"><i class="mdi ${icon} price-pin__type" aria-hidden="true"></i><span>${label}</span></span>`,
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

function formatArea(area: number): string {
    if (!area || area <= 0) return "—";
    return new Intl.NumberFormat("en-CA", { maximumFractionDigits: 0 }).format(area);
}

function formatBath(count: number): string {
    if (!count || count <= 0) return "—";
    return Number.isInteger(count) ? `${count}` : count.toFixed(1);
}

function formatBuildingType(buildingType: number): string {
    switch (buildingType) {
        case 2:
            return "2-plex";
        case 3:
            return "3-plex";
        case 19:
            return "4-plex";
        case 1:
        default:
            return "House";
    }
}

type SheetPanePosition = "previous" | "current" | "next";
type SheetPane = {
    key: string;
    pin: ListingMapPin;
    position: SheetPanePosition;
};
const selectedListingIndex = computed(() => {
    if (!selectedPin.value) return -1;
    const key = listingKey(selectedPin.value.board, selectedPin.value.mls);
    return visiblePins.value.findIndex((pin) => listingKey(pin.board, pin.mls) === key);
});
const canSwipeListings = computed(() => props.mobileSheet && visiblePins.value.length > 1);
const sheetCarouselPanes = computed<SheetPane[]>(() => {
    if (!selectedPin.value) return [];
    const index = selectedListingIndex.value;
    const pins = visiblePins.value;
    if (pins.length < 2 || index < 0) {
        return [{
            key: `${listingKey(selectedPin.value.board, selectedPin.value.mls)}-current`,
            pin: selectedPin.value,
            position: "current",
        }];
    }

    const previous = pins[(index - 1 + pins.length) % pins.length];
    const current = pins[index];
    const next = pins[(index + 1) % pins.length];
    return [
        { key: `${listingKey(previous.board, previous.mls)}-previous`, pin: previous, position: "previous" },
        { key: `${listingKey(current.board, current.mls)}-current`, pin: current, position: "current" },
        { key: `${listingKey(next.board, next.mls)}-next`, pin: next, position: "next" },
    ];
});
const sheetSwipeClasses = computed(() => ({
    "mobile-listing-sheet__swipe-track--dragging": swipeState.value === "horizontal",
    "mobile-listing-sheet__swipe-track--settling": swipeTransition.value,
}));
const sheetSwipeTrackStyle = computed(() => {
    const base = sheetCarouselPanes.value.length > 1 ? "-100%" : "0px";
    return {
        transform: `translate3d(calc(${base} + ${swipeOffsetX.value}px), 0, 0)`,
    };
});

function selectedBathLabel(pin: ListingMapPin): string {
    return formatBath(pin.bathroomCount);
}

function selectedAreaLabel(pin: ListingMapPin): string {
    return formatArea(pin.interiorAreaSqft);
}

type CloseSheetOptions = {
    immediate?: boolean;
};

function openSheet(pin: ListingMapPin): void {
    resetSheetDrag();
    resetSheetSwipe();
    animateMobileSheet.value = true;
    selectSheetPin(pin);
}

function closeSheet(options: CloseSheetOptions = {}): void {
    if (options.immediate) {
        animateMobileSheet.value = false;
    }
    resetSheetSwipe();
    selectedPin.value = null;
    clearSelectedPin();
    if (options.immediate) {
        void nextTick(() => {
            animateMobileSheet.value = true;
        });
    }
}

function selectSheetPin(pin: ListingMapPin): void {
    selectedPin.value = pin;
    markSelectedPin(pin.board, pin.mls);
}

function pinAddress(pin: ListingMapPin): { street: string; locality: string } {
    return parseAddress(pin.address);
}

function pinPhotoUrl(pin: ListingMapPin): string {
    return pin.photoUrl ?? "";
}

function showPinPhoto(pin: ListingMapPin): boolean {
    return pinPhotoUrl(pin) !== "" && !failedPhotoKeys.value.has(listingKey(pin.board, pin.mls));
}

function pinPhotoAlt(pin: ListingMapPin): string {
    return `Listing photo for ${pinAddress(pin).street}`;
}

function pinPropertyType(pin: ListingMapPin): string {
    return formatBuildingType(pin.buildingType);
}

function pinCommuteTier(pin: ListingMapPin): "fast" | "mid" | "slow" | "unknown" {
    return commuteTier(pin.commuteSecondsDowntown ?? null);
}

function pinCommuteLabel(pin: ListingMapPin): string {
    return pin.commuteSecondsDowntown != null
        ? `${Math.round(pin.commuteSecondsDowntown / 60)} min`
        : "—";
}

function pinDirectionsUrl(pin: ListingMapPin): string | null {
    return commuteMapUrl(pin.address);
}

function clearSwipeTimer(): void {
    if (swipeTimer == null) return;
    window.clearTimeout(swipeTimer);
    swipeTimer = null;
}

function resetSheetSwipe(): void {
    clearSwipeTimer();
    swipePointerId = null;
    swipeCaptureEl = null;
    swipeOffsetX.value = 0;
    swipeTransition.value = false;
    swipeState.value = "idle";
    swipeVelocityX = 0;
}

function settleSwipeBack(): void {
    clearSwipeTimer();
    swipeTransition.value = true;
    swipeOffsetX.value = 0;
    swipeTimer = window.setTimeout(() => {
        swipeTimer = null;
        swipeTransition.value = false;
        swipeState.value = "idle";
    }, 220);
}

function animateSwipeToListing(direction: 1 | -1): void {
    const pins = visiblePins.value;
    const currentIndex = selectedListingIndex.value;
    if (pins.length < 2 || currentIndex < 0) {
        settleSwipeBack();
        return;
    }

    const width = swipeStageEl.value?.offsetWidth ?? window.innerWidth;
    const nextIndex = (currentIndex + direction + pins.length) % pins.length;
    const nextPin = pins[nextIndex];

    clearSwipeTimer();
    swipeTransition.value = true;
    swipeOffsetX.value = -direction * width;
    swipeTimer = window.setTimeout(() => {
        swipeTimer = null;
        swipeTransition.value = false;
        selectSheetPin(nextPin);
        swipeOffsetX.value = 0;
        swipeState.value = "idle";
    }, 240);
}

function isInteractiveSwipeTarget(target: EventTarget | null): boolean {
    return target instanceof Element && !!target.closest(
        "a, button, input, textarea, select, [role='button'], .mobile-sheet-grip",
    );
}

function onSheetSwipePointerDown(event: PointerEvent): void {
    if (
        event.button !== 0 ||
        !canSwipeListings.value ||
        swipeTransition.value ||
        isInteractiveSwipeTarget(event.target)
    ) {
        return;
    }
    const target = event.currentTarget;
    if (!(target instanceof HTMLElement)) return;

    clearSwipeTimer();
    swipePointerId = event.pointerId;
    swipeStartX = event.clientX;
    swipeStartY = event.clientY;
    swipeLastX = event.clientX;
    swipeLastAt = performance.now();
    swipeVelocityX = 0;
    swipeOffsetX.value = 0;
    swipeTransition.value = false;
    swipeState.value = "pending";
}

function onSheetSwipePointerMove(event: PointerEvent): void {
    if (swipePointerId !== event.pointerId || swipeState.value === "idle") return;

    const dx = event.clientX - swipeStartX;
    const dy = event.clientY - swipeStartY;
    const absX = Math.abs(dx);
    const absY = Math.abs(dy);

    if (swipeState.value === "pending" && Math.max(absX, absY) > 8) {
        swipeState.value = absX > absY * 1.15 ? "horizontal" : "vertical";
        if (swipeState.value === "horizontal") {
            const target = event.currentTarget;
            if (target instanceof HTMLElement) {
                target.setPointerCapture(event.pointerId);
                swipeCaptureEl = target;
            }
        }
    }

    if (swipeState.value !== "horizontal") return;

    const now = performance.now();
    const dt = Math.max(1, now - swipeLastAt);
    swipeVelocityX = (event.clientX - swipeLastX) / dt;
    swipeLastX = event.clientX;
    swipeLastAt = now;
    swipeOffsetX.value = dx;
    event.preventDefault();
}

function onSheetSwipePointerUp(event: PointerEvent): void {
    if (swipePointerId !== event.pointerId) return;
    const target = event.currentTarget;
    const captureEl = swipeCaptureEl ?? (target instanceof HTMLElement ? target : null);
    if (captureEl?.hasPointerCapture(event.pointerId)) {
        captureEl.releasePointerCapture(event.pointerId);
    }
    swipePointerId = null;
    swipeCaptureEl = null;

    if (swipeState.value !== "horizontal") {
        swipeState.value = "idle";
        swipeOffsetX.value = 0;
        return;
    }

    const width = swipeStageEl.value?.offsetWidth ?? window.innerWidth;
    const threshold = Math.min(120, Math.max(72, width * 0.22));
    const fastSwipe = Math.abs(swipeVelocityX) > 0.55 && Math.abs(swipeOffsetX.value) > 28;
    const shouldNavigate = Math.abs(swipeOffsetX.value) > threshold || fastSwipe;

    if (shouldNavigate) {
        animateSwipeToListing(swipeOffsetX.value < 0 ? 1 : -1);
        return;
    }

    settleSwipeBack();
}

function onSheetSwipePointerCancel(event: PointerEvent): void {
    if (swipePointerId !== event.pointerId) return;
    swipePointerId = null;
    swipeCaptureEl = null;
    settleSwipeBack();
}

function onPinPhotoError(pin: ListingMapPin): void {
    failedPhotoKeys.value = new Set([
        ...failedPhotoKeys.value,
        listingKey(pin.board, pin.mls),
    ]);
}

function toggleSheetFavorite(pin: ListingMapPin): void {
    emit("toggle-favorite", {
        board: pin.board,
        mls: pin.mls,
        isFavorite: pin.isFavorite,
    });
}

function favButtonHtml(isFavorite: boolean): string {
    const icon = isFavorite ? "mdi-heart" : "mdi-heart-outline";
    const cls = isFavorite ? " map-popup__fav--active" : "";
    const label = isFavorite ? "Remove from favorites" : "Add to favorites";
    return `<button type="button" class="map-popup__fav${cls}" aria-pressed="${isFavorite}" aria-label="${label}"><i class="mdi ${icon}" aria-hidden="true"></i></button>`;
}

function popupHtml(pin: ListingMapPin): string {
    const { street, locality } = parseAddress(pin.address);
    const streetEsc = escapeHtml(street);
    const localityEsc = escapeHtml(locality);
    const price = escapeHtml(formatPrice(pin.currentPrice));
    const tier = commuteTier(pin.commuteSecondsDowntown);
    const expiredBadge = pin.isAvailable
        ? ""
        : `<div class="map-popup__expired">Expired listing</div>`;
    const commuteLabel =
        pin.commuteSecondsDowntown != null
            ? `${Math.round(pin.commuteSecondsDowntown / 60)} min`
            : "—";
    const slug = escapeHtml(pin.slug);
    const directionsHref = commuteMapUrl(pin.address);
    const directionsBtn = directionsHref
        ? `<a href="${escapeHtml(directionsHref)}" target="_blank" rel="noopener noreferrer" class="map-popup__btn map-popup__btn--ghost" aria-label="Directions">
                <i class="mdi mdi-navigation-variant-outline" aria-hidden="true"></i>
            </a>`
        : "";
    const localityLine = localityEsc
        ? `<div class="map-popup__locality">${localityEsc}</div>`
        : "";
    const bd = pin.bedroomCount > 0 ? `${pin.bedroomCount}` : "—";
    const ba = formatBath(pin.bathroomCount);
    const area = formatArea(pin.interiorAreaSqft);
    const buildingType = escapeHtml(formatBuildingType(pin.buildingType));
    const photoUrl = pin.photoUrl ? escapeHtml(pin.photoUrl) : "";
    const photo = photoUrl
        ? `<div class="map-popup__photo"><img src="${photoUrl}" alt="Listing photo" loading="lazy" decoding="async" referrerpolicy="no-referrer"></div>`
        : "";
    const photoClass = photoUrl ? "" : " map-popup--no-photo";
    const actionsClass = photoUrl
        ? "map-popup__top-actions"
        : "map-popup__top-actions map-popup__top-actions--plain";
    return `
        <div class="map-popup${photoClass}">
            ${photo}
            <div class="${actionsClass}">
                ${favButtonHtml(pin.isFavorite)}
                <button type="button" class="map-popup__close" aria-label="Close"><i class="mdi mdi-close" aria-hidden="true"></i></button>
            </div>
            <div class="map-popup__body">
                <div class="map-popup__top">
                    <div class="map-popup__header">
                        <div class="map-popup__price">${price}</div>
                        <div class="map-popup__street">${streetEsc}</div>
                        ${localityLine}
                        ${expiredBadge}
                    </div>
                </div>
                <div class="map-popup__stats">
                    <div class="map-popup__stat">
                        <i class="mdi mdi-home-outline map-popup__stat-icon" aria-hidden="true"></i>
                        <div class="map-popup__stat-key">${buildingType}</div>
                    </div>
                    <div class="map-popup__stat">
                        <i class="mdi mdi-bed-outline map-popup__stat-icon" aria-hidden="true"></i>
                        <div class="map-popup__stat-value">${bd}<span class="map-popup__stat-unit">bd</span></div>
                    </div>
                    <div class="map-popup__stat">
                        <i class="mdi mdi-bathtub-outline map-popup__stat-icon" aria-hidden="true"></i>
                        <div class="map-popup__stat-value">${ba}<span class="map-popup__stat-unit">ba</span></div>
                    </div>
                    <div class="map-popup__stat">
                        <i class="mdi mdi-ruler-square map-popup__stat-icon" aria-hidden="true"></i>
                        <div class="map-popup__stat-value">${area}<span class="map-popup__stat-unit">ft²</span></div>
                    </div>
                </div>
                <div class="map-popup__commute-row">
                    <span class="map-popup__commute-label">Commute</span>
                    <div class="map-popup__commute map-popup__commute--${tier}">
                        <span class="map-popup__commute-dot" aria-hidden="true"></span>
                        <span class="map-popup__commute-text"><b>${commuteLabel}</b> to McGill</span>
                    </div>
                </div>
            </div>
            <div class="map-popup__actions">
                <a href="${slug}" target="_blank" rel="noopener noreferrer" class="map-popup__btn map-popup__btn--primary">
                    <span>View listing</span>
                </a>
                ${directionsBtn}
            </div>
        </div>
    `;
}

async function load() {
    if (!map.value || !cluster.value) return;
    loading.value = true;
    try {
        const mapPins = await listListingsForMap({
            ...(props.buildingTypes.length > 0 && { buildingTypes: props.buildingTypes }),
            ...(props.maxPrice != null && { maxPrice: props.maxPrice }),
            ...(props.maxCommuteSec != null && {
                maxCommuteSec: props.maxCommuteSec,
            }),
            ...(props.newWithinDays != null && {
                newWithinDays: props.newWithinDays,
            }),
            ...(props.minBedrooms != null && { minBedrooms: props.minBedrooms }),
            ...(props.minBathrooms != null && { minBathrooms: props.minBathrooms }),
            ...(props.minInteriorAreaSqft != null && {
                minInteriorAreaSqft: props.minInteriorAreaSqft,
            }),
            ...(props.favoritesOnly && { favoritesOnly: true }),
            ...(props.includeExpired && { includeExpired: true }),
        });
        clearHighlight();
        closeSheet();
        visiblePins.value = mapPins;
        cluster.value.clearLayers();
        markersByKey.clear();
        const markers: L.Marker[] = [];
        for (const pin of mapPins) {
            const m = L.marker([pin.latitude, pin.longitude], {
                icon: pricePillIcon(
                    pin.currentPrice,
                    pin.commuteSecondsDowntown,
                    pin.buildingType,
                ),
                riseOnHover: true,
            }) as L.Marker & { _data?: MarkerData };
            m._data = {
                board: pin.board,
                mls: pin.mls,
                price: pin.currentPrice,
                commuteSec: pin.commuteSecondsDowntown,
                isFavorite: pin.isFavorite,
            };
            if (!props.mobileSheet) {
                m.bindPopup(popupHtml(pin), { closeButton: false });
                m.on("popupopen", () => wireFavButton(m, pin));
            }
            m.on("click", (event) => {
                L.DomEvent.stopPropagation(event);
                emit("pin-click", { board: pin.board, mls: pin.mls });
                if (props.mobileSheet) openSheet(pin);
            });
            markers.push(m);
            markersByKey.set(listingKey(pin.board, pin.mls), m);
        }
        cluster.value.addLayers(markers);
        emit("update:count", markers.length);
        if (markers.length > 0 && (!hasFitBounds || forceRefit)) {
            const bounds = L.latLngBounds(
                markers.map((m) => m.getLatLng()),
            );
            map.value.fitBounds(bounds, { padding: [40, 40], maxZoom: 14 });
            hasFitBounds = true;
        }
        forceRefit = false;
    } catch (e: any) {
        emit(
            "error",
            e?.response?.data?.error ??
            e?.message ??
            "failed to load map listings",
        );
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
        attributionControl: false,
        preferCanvas: true,
    });
    map.value.on("click", () => {
        if (props.mobileSheet) closeSheet();
    });
    L.control
        .attribution({ prefix: false, position: "bottomleft" })
        .addAttribution(
            '&copy; <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions" target="_blank" rel="noopener">CARTO</a>',
        )
        .addTo(map.value);
    map.value.createPane("hexPane");
    const hexPane = map.value.getPane("hexPane");
    if (hexPane) {
        hexPane.style.zIndex = "350";
        hexPane.style.pointerEvents = "none";
    }
    L.tileLayer(
        "https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png",
        {
            subdomains: "abcd",
            maxZoom: 19,
            updateWhenZooming: false,
            updateWhenIdle: false,
            keepBuffer: 4,
        },
    ).addTo(map.value);
    cluster.value = L.markerClusterGroup({
        showCoverageOnHover: false,
        spiderfyOnMaxZoom: true,
        chunkedLoading: true,
        maxClusterRadius: mobile.value ? 58 : 65,
        disableClusteringAtZoom: 15,
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
    debouncedLoad.cancel();
    resetSheetSwipe();
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
    () => props.favoritesOnly,
    (v) => {
        if (v) forceRefit = true;
    },
);

const debouncedLoad = debounce(() => load(), 250);

watch(
    () => [
        props.buildingTypes.join(","),
        props.maxPrice,
        props.maxCommuteSec,
        props.newWithinDays,
        props.minBedrooms,
        props.minBathrooms,
        props.minInteriorAreaSqft,
        props.favoritesOnly,
        props.includeExpired,
    ],
    debouncedLoad,
);

watch(
    () => props.mobileSheet,
    () => {
        closeSheet();
        debouncedLoad();
    },
);

watch(mobile, () => {
    debouncedLoad();
});

watch(loading, (v) => emit("update:loading", v));

function wireFavButton(
    m: L.Marker & { _data?: MarkerData },
    pin: ListingMapPin,
): void {
    const el = m.getPopup()?.getElement();
    const btn = el?.querySelector(".map-popup__fav") as HTMLButtonElement | null;
    if (!btn) return;
    btn.onclick = (ev) => {
        ev.stopPropagation();
        emit("toggle-favorite", {
            board: pin.board,
            mls: pin.mls,
            isFavorite: pin.isFavorite,
        });
    };
    const closeBtn = el?.querySelector(
        ".map-popup__close",
    ) as HTMLButtonElement | null;
    if (closeBtn) {
        closeBtn.onclick = (ev) => {
            ev.stopPropagation();
            m.closePopup();
        };
    }
}

function setFavorite(board: number, mls: number, value: boolean): void {
    const m = markersByKey.get(listingKey(board, mls)) as
        | (L.Marker & { _data?: MarkerData })
        | undefined;
    if (!m) return;
    if (m._data) m._data.isFavorite = value;
    visiblePins.value = visiblePins.value.map((pin) =>
        pin.board === board && pin.mls === mls
            ? { ...pin, isFavorite: value }
            : pin,
    );
    if (
        selectedPin.value &&
        selectedPin.value.board === board &&
        selectedPin.value.mls === mls
    ) {
        selectedPin.value = { ...selectedPin.value, isFavorite: value };
    }
    const el = m.getPopup()?.getElement();
    const btn = el?.querySelector(".map-popup__fav") as HTMLButtonElement | null;
    if (!btn) return;
    btn.classList.toggle("map-popup__fav--active", value);
    btn.setAttribute("aria-pressed", String(value));
    btn.setAttribute(
        "aria-label",
        value ? "Remove from favorites" : "Add to favorites",
    );
    const icon = btn.querySelector("i");
    if (icon) icon.className = `mdi ${value ? "mdi-heart" : "mdi-heart-outline"}`;
}

function focusListing(board: number, mls: number): boolean {
    const marker = markersByKey.get(listingKey(board, mls));
    if (!marker || !map.value || !cluster.value) return false;
    cluster.value.zoomToShowLayer(marker, () => {
        marker.openPopup();
    });
    return true;
}

let highlightedEl: HTMLElement | null = null;

function markSelectedPin(board: number, mls: number): void {
    clearSelectedPin();
    const marker = markersByKey.get(listingKey(board, mls));
    if (!marker || !cluster.value) return;
    const visible = cluster.value.getVisibleParent(marker) ?? marker;
    const el = (visible as L.Marker | L.MarkerCluster).getElement?.();
    if (!el) return;
    el.classList.add("price-pin--selected");
    selectedMarkerEl = el;
}

function clearSelectedPin(): void {
    if (selectedMarkerEl) {
        selectedMarkerEl.classList.remove("price-pin--selected");
        selectedMarkerEl = null;
    }
}

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

defineExpose({ focusListing, highlightListing, clearHighlight, setFavorite, closeSheet });
</script>

<template>
    <div class="listings-map-wrap">
        <div ref="mapEl" class="listings-map" />
        <div class="listings-map-legend" aria-label="Transit commute time legend">
            <div class="listings-map-legend__title">Commute to downtown</div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch listings-map-legend__swatch--fast" />
                &lt; 30 min
            </div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch listings-map-legend__swatch--mid" />
                30–60 min
            </div>
            <div class="listings-map-legend__row">
                <span class="listings-map-legend__swatch listings-map-legend__swatch--slow" />
                &gt; 60 min
            </div>
        </div>

        <Transition v-if="props.mobileSheet" name="mobile-sheet-transition" :css="animateMobileSheet">
            <div
                v-if="selectedPin"
                class="mobile-listing-sheet"
                :class="sheetDragClasses"
                :style="sheetDragStyle"
                role="dialog"
                aria-label="Listing details"
            >
                <button
                    type="button"
                    class="mobile-sheet-grip"
                    aria-label="Drag down to close listing details"
                    @pointerdown="onDragPointerDown"
                    @pointermove="onDragPointerMove"
                    @pointerup="onDragPointerUp"
                    @pointercancel="onDragPointerCancel"
                />

                <div
                    ref="swipeStageEl"
                    class="mobile-listing-sheet__swipe-viewport"
                    @pointerdown="onSheetSwipePointerDown"
                    @pointermove="onSheetSwipePointerMove"
                    @pointerup="onSheetSwipePointerUp"
                    @pointercancel="onSheetSwipePointerCancel"
                >
                    <div
                        class="mobile-listing-sheet__swipe-track"
                        :class="sheetSwipeClasses"
                        :style="sheetSwipeTrackStyle"
                    >
                        <article
                            v-for="pane in sheetCarouselPanes"
                            :key="pane.key"
                            class="mobile-listing-sheet__swipe-pane"
                            :aria-hidden="pane.position !== 'current'"
                        >
                            <div
                                class="mobile-listing-sheet__photo"
                                :class="{ 'mobile-listing-sheet__photo--placeholder': !showPinPhoto(pane.pin) }"
                            >
                                <img
                                    v-if="showPinPhoto(pane.pin)"
                                    :src="pinPhotoUrl(pane.pin)"
                                    :alt="pinPhotoAlt(pane.pin)"
                                    loading="eager"
                                    decoding="async"
                                    referrerpolicy="no-referrer"
                                    @error="onPinPhotoError(pane.pin)"
                                >
                                <div v-else class="mobile-listing-sheet__placeholder" role="img" aria-label="No listing photos available">
                                    <v-icon size="34">mdi-image-off-outline</v-icon>
                                    <span>No photos</span>
                                </div>
                            </div>
                            <div class="mobile-listing-sheet__top-actions">
                                <button
                                    type="button"
                                    class="mobile-listing-sheet__iconbtn"
                                    :class="{ 'mobile-listing-sheet__iconbtn--active': pane.pin.isFavorite }"
                                    :aria-pressed="pane.pin.isFavorite"
                                    :aria-label="pane.pin.isFavorite ? 'Remove from favorites' : 'Add to favorites'"
                                    @click="toggleSheetFavorite(pane.pin)"
                                >
                                    <v-icon size="20">{{ pane.pin.isFavorite ? "mdi-heart" : "mdi-heart-outline" }}</v-icon>
                                </button>
                                <button
                                    type="button"
                                    class="mobile-listing-sheet__iconbtn"
                                    aria-label="Close listing details"
                                    @click="closeSheet()"
                                >
                                    <v-icon size="21">mdi-close</v-icon>
                                </button>
                            </div>
                            <div class="mobile-listing-sheet__body">
                                <div class="mobile-listing-sheet__top">
                                    <div class="mobile-listing-sheet__header">
                                        <div class="mobile-listing-sheet__price">
                                            {{ formatPrice(pane.pin.currentPrice) }}
                                        </div>
                                        <div class="mobile-listing-sheet__street">
                                            {{ pinAddress(pane.pin).street }}
                                        </div>
                                        <div v-if="pinAddress(pane.pin).locality" class="mobile-listing-sheet__locality">
                                            {{ pinAddress(pane.pin).locality }}
                                        </div>
                                        <div v-if="!pane.pin.isAvailable" class="mobile-listing-sheet__expired">
                                            Expired listing
                                        </div>
                                    </div>
                                </div>

                                <div class="mobile-listing-sheet__stats">
                                    <div class="mobile-listing-sheet__stat">
                                        <v-icon size="18">mdi-home-outline</v-icon>
                                        <span class="mobile-listing-sheet__stat-key">{{ pinPropertyType(pane.pin) }}</span>
                                    </div>
                                    <div class="mobile-listing-sheet__stat">
                                        <v-icon size="18">mdi-bed-outline</v-icon>
                                        <span class="mobile-listing-sheet__stat-value">
                                            {{ pane.pin.bedroomCount > 0 ? pane.pin.bedroomCount : "—" }}<small>bd</small>
                                        </span>
                                    </div>
                                    <div class="mobile-listing-sheet__stat">
                                        <v-icon size="18">mdi-bathtub-outline</v-icon>
                                        <span class="mobile-listing-sheet__stat-value">
                                            {{ selectedBathLabel(pane.pin) }}<small>ba</small>
                                        </span>
                                    </div>
                                    <div class="mobile-listing-sheet__stat">
                                        <v-icon size="18">mdi-ruler-square</v-icon>
                                        <span class="mobile-listing-sheet__stat-value">
                                            {{ selectedAreaLabel(pane.pin) }}<small>ft²</small>
                                        </span>
                                    </div>
                                </div>

                                <div class="mobile-listing-sheet__commute-row">
                                    <span class="mobile-listing-sheet__commute-label">Commute</span>
                                    <span
                                        class="mobile-listing-sheet__commute"
                                        :class="`mobile-listing-sheet__commute--${pinCommuteTier(pane.pin)}`"
                                    >
                                        <span class="mobile-listing-sheet__commute-dot" />
                                        <span class="mobile-listing-sheet__commute-text">
                                            <b>{{ pinCommuteLabel(pane.pin) }}</b> to McGill
                                        </span>
                                    </span>
                                </div>
                            </div>

                            <div class="mobile-listing-sheet__actions">
                                <a
                                    :href="pane.pin.slug"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="mobile-listing-sheet__btn mobile-listing-sheet__btn--primary"
                                >
                                    View listing
                                </a>
                                <a
                                    v-if="pinDirectionsUrl(pane.pin)"
                                    :href="pinDirectionsUrl(pane.pin) ?? undefined"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    class="mobile-listing-sheet__btn mobile-listing-sheet__btn--ghost"
                                    aria-label="Directions"
                                >
                                    <v-icon size="20">mdi-navigation-variant-outline</v-icon>
                                </a>
                            </div>
                        </article>
                    </div>
                </div>
            </div>
        </Transition>
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

.listings-map-legend {
    position: absolute;
    bottom: 16px;
    right: 8px;
    z-index: 500;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 10px;
    background-color: rgba(var(--v-theme-surface), 0.6);
    color: rgba(var(--v-theme-on-surface), 0.92);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 6px;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    font-size: 11px;
    line-height: 1.2;
    pointer-events: none;
    box-shadow: 0 2px 6px rgba(var(--v-theme-shadow), 0.25);
}

.listings-map-legend__title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: rgba(var(--v-theme-on-surface), 0.65);
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
    box-shadow: 0 0 0 1px rgba(var(--v-theme-shadow), 0.3);
}

.listings-map-legend__swatch--fast {
    background-color: rgb(var(--v-theme-commute-fast));
}

.listings-map-legend__swatch--mid {
    background-color: rgb(var(--v-theme-commute-mid));
}

.listings-map-legend__swatch--slow {
    background-color: rgb(var(--v-theme-commute-slow));
}

.mobile-listing-sheet {
    position: fixed;
    right: 11px;
    bottom: 0;
    left: 11px;
    z-index: 2000;
    display: flex;
    flex-direction: column;
    width: 100%;
    height: min(640px, calc(100dvh - 72px));
    max-width: calc(100vw - 22px);
    max-height: calc(100dvh - 72px);
    --mobile-listing-sheet-bg: 38, 41, 37;
    color: rgb(var(--v-theme-on-surface));
    background: rgb(var(--mobile-listing-sheet-bg));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 22px 22px 0 0;
    overflow: hidden;
    box-shadow:
        0 -22px 70px -28px rgba(var(--v-theme-shadow), 0.95),
        inset 0 1px 0 rgba(var(--v-theme-on-surface), 0.04);
}

.mobile-listing-sheet__swipe-viewport {
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
    overflow: hidden;
    touch-action: pan-y;
}

.mobile-listing-sheet__swipe-track {
    height: 100%;
    display: flex;
    min-height: 0;
    transform: translate3d(0, 0, 0);
    will-change: transform;
}

.mobile-listing-sheet__swipe-track--dragging {
    transition: none;
}

.mobile-listing-sheet__swipe-track--settling {
    transition: transform 240ms cubic-bezier(0.2, 0.8, 0.2, 1);
}

.mobile-listing-sheet__swipe-pane {
    position: relative;
    flex: 0 0 100%;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    touch-action: pan-y;
}

.mobile-listing-sheet__photo {
    width: 100%;
    height: clamp(172px, 34dvh, 230px);
    flex: 0 0 auto;
    overflow: hidden;
    background: rgba(var(--v-theme-on-surface), 0.05);
}

.mobile-listing-sheet__photo--placeholder {
    background:
        color-mix(in srgb, rgb(var(--mobile-listing-sheet-bg)) 88%, rgb(var(--v-theme-primary)) 12%);
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
}

.mobile-listing-sheet__photo img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.mobile-listing-sheet__placeholder {
    display: flex;
    width: 100%;
    height: 100%;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: rgba(var(--v-theme-on-surface), 0.46);
}

.mobile-listing-sheet__placeholder span {
    font-size: 13px;
    font-weight: 800;
    letter-spacing: 0.08em;
    line-height: 1;
    text-transform: uppercase;
}

.mobile-listing-sheet__body {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
    padding: 22px 22px 6px;
    touch-action: pan-y;
}

.mobile-listing-sheet__top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.mobile-listing-sheet__header {
    flex: 1 1 auto;
    min-width: 0;
}

.mobile-listing-sheet__price {
    margin-bottom: 10px;
    font-size: 30px;
    font-weight: 800;
    letter-spacing: 0;
    line-height: 1;
    font-variant-numeric: tabular-nums;
}

.mobile-listing-sheet__street {
    font-size: 22px;
    font-weight: 800;
    letter-spacing: 0;
    line-height: 1.14;
    overflow-wrap: break-word;
    word-break: normal;
}

.mobile-listing-sheet__locality {
    margin-top: 4px;
    color: rgba(var(--v-theme-on-surface), 0.52);
    font-size: 16px;
    font-weight: 500;
    line-height: 1.25;
}

.mobile-listing-sheet__expired {
    display: inline-flex;
    align-items: center;
    margin-top: 10px;
    padding: 4px 10px;
    border-radius: 999px;
    border: 1px solid rgba(var(--v-theme-warning), 0.32);
    background: rgba(var(--v-theme-warning), 0.09);
    color: rgb(var(--v-theme-warning));
    font-size: 12px;
    font-weight: 700;
}

.mobile-listing-sheet__top-actions {
    position: absolute;
    top: 16px;
    right: 16px;
    z-index: 2;
    display: flex;
    align-items: center;
    gap: 8px;
}

.mobile-listing-sheet__iconbtn {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 999px;
    background: rgba(var(--mobile-listing-sheet-bg), 0.72);
    color: rgba(var(--v-theme-on-surface), 0.68);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    box-shadow: 0 4px 14px rgba(var(--v-theme-shadow), 0.28);
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease, transform 60ms ease;
}

.mobile-listing-sheet__iconbtn:hover {
    border-color: rgba(var(--v-theme-on-surface), 0.2);
    background-color: rgba(var(--mobile-listing-sheet-bg), 0.9);
    color: rgba(var(--v-theme-on-surface), 0.9);
}

.mobile-listing-sheet__iconbtn--active {
    color: rgb(var(--v-theme-error));
    border-color: rgba(var(--v-theme-error), 0.28);
    background-color: rgba(var(--v-theme-error), 0.14);
}

.mobile-listing-sheet__iconbtn--active:hover {
    color: rgb(var(--v-theme-error));
    border-color: rgba(var(--v-theme-error), 0.38);
    background-color: rgba(var(--v-theme-error), 0.18);
}

.mobile-listing-sheet__iconbtn:active {
    transform: translateY(1px);
}

.mobile-listing-sheet__stats {
    display: flex;
    align-items: stretch;
    margin: 20px 0 17px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 16px;
    background: rgba(var(--v-theme-on-surface), 0.03);
    overflow: hidden;
}

.mobile-listing-sheet__stat {
    flex: 1 1 0;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 7px;
    padding: 12px 6px 11px;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    color: rgba(var(--v-theme-on-surface), 0.52);
    text-align: center;
}

.mobile-listing-sheet__stat:first-child {
    border-left: 0;
}

.mobile-listing-sheet__stat-key {
    color: rgba(var(--v-theme-on-surface), 0.34);
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
    line-height: 1;
    text-transform: uppercase;
}

.mobile-listing-sheet__stat-value {
    color: rgb(var(--v-theme-on-surface));
    font-size: 17px;
    font-weight: 800;
    line-height: 1;
    font-variant-numeric: tabular-nums;
}

.mobile-listing-sheet__stat-value small {
    margin-left: 2px;
    color: rgba(var(--v-theme-on-surface), 0.38);
    font-size: 12px;
    font-weight: 600;
}

.mobile-listing-sheet__commute-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}

.mobile-listing-sheet__commute-label {
    flex: 0 0 auto;
    color: rgba(var(--v-theme-on-surface), 0.34);
    font-size: 12px;
    font-weight: 800;
    letter-spacing: 0.12em;
    line-height: 30px;
    text-transform: uppercase;
}

.mobile-listing-sheet__commute {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    max-width: 210px;
    height: 36px;
    padding: 0 16px 0 12px;
    border-radius: 999px;
    --commute-accent: var(--v-theme-on-surface);
    background: color-mix(in srgb, rgb(var(--commute-accent)) 13%, transparent);
    border: 1px solid color-mix(in srgb, rgb(var(--commute-accent)) 38%, transparent);
    color: color-mix(in srgb, rgb(var(--commute-accent)) 62%, rgb(var(--v-theme-on-surface)));
    white-space: nowrap;
}

.mobile-listing-sheet__commute--fast {
    --commute-accent: var(--v-theme-commute-fast);
}

.mobile-listing-sheet__commute--mid {
    --commute-accent: var(--v-theme-commute-mid);
}

.mobile-listing-sheet__commute--slow {
    --commute-accent: var(--v-theme-commute-slow);
}

.mobile-listing-sheet__commute-dot {
    width: 10px;
    height: 10px;
    flex: 0 0 auto;
    border-radius: 999px;
    background: rgb(var(--commute-accent));
    box-shadow: 0 0 0 4px color-mix(in srgb, rgb(var(--commute-accent)) 18%, transparent);
}

.mobile-listing-sheet__commute-text {
    min-width: 0;
    overflow: hidden;
    font-size: 16px;
    font-weight: 700;
    line-height: 1;
    text-overflow: ellipsis;
}

.mobile-listing-sheet__commute-text b {
    font-weight: 900;
    font-variant-numeric: tabular-nums;
    color: color-mix(in srgb, rgb(var(--commute-accent)) 80%, rgb(var(--v-theme-on-surface)));
}

.mobile-listing-sheet__actions {
    display: flex;
    gap: 12px;
    padding: 18px 22px calc(18px + env(safe-area-inset-bottom, 0px));
}

.mobile-listing-sheet__btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 60px;
    border-radius: 999px;
    text-decoration: none;
    font-size: 18px;
    font-weight: 800;
    transition: background-color 140ms ease, border-color 140ms ease, transform 60ms ease, box-shadow 140ms ease;
}

.mobile-listing-sheet__btn:active {
    transform: translateY(1px);
}

.mobile-listing-sheet__btn--primary {
    flex: 1 1 auto;
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    box-shadow: 0 8px 24px -10px rgba(var(--v-theme-primary), 0.8);
}

.mobile-listing-sheet__btn--ghost {
    flex: 0 0 auto;
    width: 60px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    background: rgba(var(--v-theme-on-surface), 0.05);
    color: rgb(var(--v-theme-on-surface));
}

@media (max-width: 959.98px) {
    .listings-map-legend {
        display: none;
    }
}

@media (max-width: 899px) {
    .mobile-sheet-transition-enter-active,
    .mobile-sheet-transition-leave-active {
        transition: opacity 180ms ease, transform 260ms cubic-bezier(0.22, 0.7, 0.3, 1);
    }

    .mobile-sheet-transition-enter-from,
    .mobile-sheet-transition-leave-to {
        opacity: 0;
        transform: translateY(100%);
    }
}
</style>

<style>
.leaflet-container {
    background: rgb(var(--v-theme-map-bg));
}

.price-pin {
    background: transparent;
    border: 0;
    --pin-ring: rgba(var(--v-theme-on-surface), 0.18);
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
    color: rgba(var(--v-theme-on-surface), 0.85);
    text-shadow:
        0 0 4px rgba(var(--v-theme-shadow), 0.85),
        0 0 8px rgba(var(--v-theme-shadow), 0.6),
        0 1px 1px rgba(var(--v-theme-shadow), 0.9);
    cursor: help;
    transition: transform 120ms ease, color 120ms ease;
}

.downtown-target:hover .downtown-target__star {
    color: rgb(var(--v-theme-warning));
    transform: scale(1.15);
}

.leaflet-tooltip.downtown-target__tooltip {
    background: rgba(var(--v-theme-popup-overlay), 0.82);
    color: rgba(var(--v-theme-on-surface), 0.92);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.02em;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow: 0 2px 6px rgba(var(--v-theme-shadow), 0.35);
    white-space: nowrap;
}

.leaflet-tooltip.downtown-target__tooltip::before {
    border-top-color: rgba(var(--v-theme-popup-overlay), 0.82);
}

.price-pin--fast {
    --pin-ring: rgba(var(--v-theme-commute-fast), 0.9);
}

.price-pin--mid {
    --pin-ring: rgba(var(--v-theme-commute-mid), 0.9);
}

.price-pin--slow {
    --pin-ring: rgba(var(--v-theme-commute-slow), 0.9);
}

.price-pin--unknown {
    --pin-ring: rgba(var(--v-theme-on-surface), 0.22);
}

.price-pin__label {
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    transform: translate(-50%, -50%);
    padding: 4px 10px 4px 9px;
    border-radius: 999px;
    background: rgba(var(--v-theme-popup-overlay), 0.88);
    color: rgb(var(--v-theme-on-surface));
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.01em;
    line-height: 1.15;
    white-space: nowrap;
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow:
        0 4px 10px rgba(var(--v-theme-shadow), 0.45),
        0 1px 2px rgba(var(--v-theme-shadow), 0.35),
        inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.06),
        inset 0 1px 0 rgba(var(--v-theme-on-surface), 0.08);
    cursor: pointer;
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.price-pin__type {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 14px;
    min-width: 14px;
    font-size: 13px;
    line-height: 1;
    color: rgba(var(--v-theme-on-surface), 0.72);
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
        0 6px 14px rgba(var(--v-theme-shadow), 0.55),
        0 1px 2px rgba(var(--v-theme-shadow), 0.4),
        inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.1);
    z-index: 1000;
}

.price-pin.price-pin--highlighted {
    z-index: 1001 !important;
}

.price-pin.price-pin--selected {
    z-index: 1002 !important;
}

.price-pin.price-pin--highlighted .price-pin__label {
    transform: translate(-50%, -50%) scale(1.18);
}

.price-pin.price-pin--selected .price-pin__label {
    transform: translate(-50%, -50%) scale(1.1);
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    box-shadow:
        0 6px 14px rgba(var(--v-theme-shadow), 0.55),
        0 1px 2px rgba(var(--v-theme-shadow), 0.35);
}

.price-pin.price-pin--selected .price-pin__type {
    color: rgba(var(--v-theme-on-primary), 0.72);
}

.price-pin.price-pin--highlighted .price-pin__label::before {
    border-color: rgb(var(--v-theme-primary));
    border-width: 2px;
    inset: -3px;
}

.price-pin.price-pin--selected .price-pin__label::before {
    border-color: transparent;
}

.price-cluster {
    background: transparent;
    border: 0;
    --cluster-ring: rgba(var(--v-theme-on-surface), 0.18);
}

.price-cluster--fast {
    --cluster-ring: rgba(var(--v-theme-commute-fast), 0.85);
}

.price-cluster--mid {
    --cluster-ring: rgba(var(--v-theme-commute-mid), 0.85);
}

.price-cluster--slow {
    --cluster-ring: rgba(var(--v-theme-commute-slow), 0.85);
}

.price-cluster--unknown {
    --cluster-ring: rgba(var(--v-theme-on-surface), 0.25);
}

.price-cluster__inner {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    width: 48px;
    height: 30px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-popup-overlay), 0.88);
    color: rgb(var(--v-theme-on-surface));
    backdrop-filter: blur(6px);
    -webkit-backdrop-filter: blur(6px);
    box-shadow:
        0 4px 10px rgba(var(--v-theme-shadow), 0.45),
        0 1px 2px rgba(var(--v-theme-shadow), 0.35),
        inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.06);
    cursor: pointer;
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.price-cluster__inner::before {
    content: "";
    position: absolute;
    inset: -2px;
    border-radius: 999px;
    border: 1.5px solid var(--cluster-ring);
    pointer-events: none;
}

.price-cluster--md .price-cluster__inner {
    width: 54px;
    height: 32px;
}

.price-cluster--lg .price-cluster__inner {
    width: 62px;
    height: 34px;
}

.price-cluster--mobile .price-cluster__inner {
    width: 56px;
    height: 34px;
}

.price-cluster--mobile .price-cluster__inner::before {
    inset: -2px;
}

.price-cluster:hover .price-cluster__inner {
    transform: scale(1.06);
    box-shadow:
        0 6px 14px rgba(var(--v-theme-shadow), 0.55),
        0 1px 2px rgba(var(--v-theme-shadow), 0.4),
        inset 0 0 0 1px rgba(var(--v-theme-on-surface), 0.1);
}

.price-cluster__count {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.01em;
    line-height: 1;
    color: rgba(var(--v-theme-on-surface), 0.96);
}

.price-cluster__label {
    margin-top: 1px;
    font-size: 7px;
    font-weight: 600;
    letter-spacing: 0;
    line-height: 1;
    color: rgba(var(--v-theme-on-surface), 0.64);
}

.price-cluster--md .price-cluster__count {
    font-size: 12px;
}

.price-cluster--lg .price-cluster__count {
    font-size: 13px;
}

.price-cluster--mobile.price-cluster--md .price-cluster__inner,
.price-cluster--mobile.price-cluster--lg .price-cluster__inner {
    width: 62px;
    height: 36px;
}

.price-cluster--mobile .price-cluster__count,
.price-cluster--mobile.price-cluster--md .price-cluster__count,
.price-cluster--mobile.price-cluster--lg .price-cluster__count {
    font-size: 13px;
}

.marker-cluster.price-pin--highlighted {
    z-index: 1001 !important;
}

.marker-cluster.price-pin--selected {
    z-index: 1002 !important;
}

.marker-cluster.price-pin--highlighted>div,
.price-cluster.price-pin--highlighted .price-cluster__inner {
    box-shadow:
        0 0 0 3px rgb(var(--v-theme-primary)),
        0 6px 14px rgba(var(--v-theme-shadow), 0.6);
    transform: scale(1.12);
    transition: transform 120ms ease, box-shadow 120ms ease;
}

.marker-cluster.price-pin--selected>div,
.price-cluster.price-pin--selected .price-cluster__inner {
    background: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    box-shadow:
        0 6px 14px rgba(var(--v-theme-shadow), 0.55),
        0 1px 2px rgba(var(--v-theme-shadow), 0.35);
    transform: scale(1.08);
    transition: transform 120ms ease, box-shadow 120ms ease, background-color 120ms ease;
}

.price-cluster.price-pin--selected .price-cluster__inner::before {
    border-color: transparent;
}

.leaflet-control-attribution {
    margin: 0 0 8px 8px !important;
    background: rgba(var(--v-theme-popup-overlay), 0.65) !important;
    color: rgba(var(--v-theme-on-surface), 0.85) !important;
    font-size: 11px !important;
    line-height: 1.5 !important;
    padding: 3px 8px !important;
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    border-radius: 6px;
}

.leaflet-control-attribution a {
    color: rgba(var(--v-theme-on-surface), 0.95) !important;
    text-decoration: underline;
}

.leaflet-popup-content-wrapper {
    background-color: rgb(var(--v-theme-popup-overlay));
    color: rgb(var(--v-theme-on-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 18px;
    box-shadow:
        0 30px 70px -18px rgba(var(--v-theme-shadow), 0.85),
        inset 0 2px 0 rgba(var(--v-theme-on-surface), 0.03);
    padding: 0;
    overflow: hidden;
}

.leaflet-popup-content {
    width: var(--map-popup-width, 332px) !important;
    margin: 0;
}

.leaflet-popup-tip {
    background-color: rgb(var(--v-theme-popup-overlay));
    border-right: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    box-shadow: none;
}

.map-popup {
    position: relative;
    width: var(--map-popup-width, 332px);
    overflow: hidden;
    font-size: 14px;
}

.map-popup__photo {
    width: 100%;
    height: 154px;
    overflow: hidden;
    background: rgba(var(--v-theme-on-surface), 0.05);
}

.map-popup__photo img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: cover;
}

.map-popup__body {
    padding: 18px 18px 6px;
}

.map-popup__top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.map-popup__header {
    flex: 1 1 auto;
    min-width: 0;
}

.map-popup--no-photo .map-popup__header {
    padding-right: 84px;
}

.map-popup__top-actions {
    position: absolute;
    top: 12px;
    right: 12px;
    z-index: 2;
    display: inline-flex;
    align-items: center;
    gap: 8px;
}

.map-popup__fav {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 999px;
    background: rgba(var(--v-theme-popup-overlay), 0.72);
    color: rgba(var(--v-theme-on-surface), 0.68);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    box-shadow: 0 4px 14px rgba(var(--v-theme-shadow), 0.28);
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease, transform 60ms ease;
}

.map-popup__fav .mdi {
    font-size: 19px;
    transform: translateY(1px);
}

.map-popup__close {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 999px;
    background: rgba(var(--v-theme-popup-overlay), 0.72);
    color: rgba(var(--v-theme-on-surface), 0.68);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    box-shadow: 0 4px 14px rgba(var(--v-theme-shadow), 0.28);
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease, transform 60ms ease;
}

.map-popup__close .mdi {
    font-size: 20px;
    transform: translateY(0.5px);
}

.map-popup__close:hover {
    border-color: rgba(var(--v-theme-on-surface), 0.2);
    background-color: rgba(var(--v-theme-popup-overlay), 0.9);
    color: rgba(var(--v-theme-on-surface), 0.9);
}

.map-popup__fav:hover {
    border-color: rgba(var(--v-theme-on-surface), 0.2);
    background-color: rgba(var(--v-theme-popup-overlay), 0.9);
    color: rgba(var(--v-theme-on-surface), 0.9);
}

.map-popup__fav--active {
    color: rgb(var(--v-theme-error));
    border-color: rgba(var(--v-theme-error), 0.28);
    background-color: rgba(var(--v-theme-error), 0.14);
}

.map-popup__fav--active:hover {
    color: rgb(var(--v-theme-error));
    border-color: rgba(var(--v-theme-error), 0.38);
    background-color: rgba(var(--v-theme-error), 0.18);
}

.map-popup__fav:active,
.map-popup__close:active {
    transform: translateY(1px);
}

.map-popup__top-actions--plain .map-popup__fav,
.map-popup__top-actions--plain .map-popup__close {
    border: 0;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.56);
    box-shadow: none;
    text-shadow: none;
    backdrop-filter: none;
    -webkit-backdrop-filter: none;
}

.map-popup__top-actions--plain .map-popup__fav:hover,
.map-popup__top-actions--plain .map-popup__close:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.86);
}

.map-popup__top-actions--plain .map-popup__fav--active {
    color: rgb(var(--v-theme-error));
}

.map-popup__top-actions--plain .map-popup__fav--active:hover {
    background-color: rgba(var(--v-theme-error), 0.08);
    color: rgb(var(--v-theme-error));
}

.map-popup__price {
    margin-bottom: 9px;
    font-size: 23px;
    font-weight: 800;
    line-height: 1;
    letter-spacing: 0;
    font-variant-numeric: tabular-nums;
    color: rgba(var(--v-theme-on-surface), 0.98);
}

.map-popup__expired {
    display: inline-flex;
    align-items: center;
    margin-top: 8px;
    padding: 3px 10px;
    border-radius: 999px;
    background-color: rgba(var(--v-theme-warning), 0.16);
    color: rgb(var(--v-theme-warning));
    border: 1px solid rgba(var(--v-theme-warning), 0.32);
    font-size: 12px;
    font-weight: 600;
}

.map-popup__street {
    font-size: 16px;
    font-weight: 700;
    line-height: 1.25;
    letter-spacing: 0;
    color: rgba(var(--v-theme-on-surface), 0.98);
    overflow-wrap: anywhere;
}

.map-popup__locality {
    margin-top: 3px;
    font-size: 13px;
    line-height: 1.3;
    color: rgba(var(--v-theme-on-surface), 0.52);
    font-weight: 500;
    overflow-wrap: anywhere;
}

.map-popup__stats {
    display: flex;
    align-items: stretch;
    margin: 16px 0 14px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    border-radius: 13px;
    background: rgba(var(--v-theme-on-surface), 0.03);
    overflow: hidden;
}

.map-popup__stat {
    flex: 1 1 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    min-width: 0;
    padding: 12px 6px 11px;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    text-align: center;
}

.map-popup__stat:first-child {
    border-left: 0;
}

.map-popup__stat-icon {
    font-size: 18px;
    color: rgba(var(--v-theme-on-surface), 0.52);
    line-height: 1;
}

.map-popup__stat-value {
    max-width: 100%;
    font-size: 14px;
    font-weight: 700;
    letter-spacing: 0;
    color: rgba(var(--v-theme-on-surface), 0.95);
    line-height: 1;
    overflow-wrap: anywhere;
}

.map-popup__stat-key {
    font-size: 10.5px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.34);
    font-weight: 600;
    line-height: 1;
}

.map-popup__stat-unit {
    margin-left: 2px;
    font-size: 12px;
    font-weight: 600;
    color: rgba(var(--v-theme-on-surface), 0.34);
}

.map-popup__commute-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    margin-bottom: 4px;
}

.map-popup__commute-label {
    font-size: 11px;
    line-height: 28px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.34);
    font-weight: 700;
    flex: 0 0 auto;
}

.map-popup__commute {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    height: 28px;
    padding: 0 12px 0 10px;
    border-radius: 999px;
    font-size: 12.5px;
    font-weight: 600;
    flex: 0 1 auto;
    min-width: 0;
    white-space: nowrap;
    --commute-accent: var(--v-theme-on-surface);
    background: color-mix(in srgb, rgb(var(--commute-accent)) 13%, transparent);
    border: 1px solid color-mix(in srgb, rgb(var(--commute-accent)) 38%, transparent);
    color: color-mix(in srgb, rgb(var(--commute-accent)) 62%, rgb(var(--v-theme-on-surface)));
}

.map-popup__commute--fast {
    --commute-accent: var(--v-theme-commute-fast);
}

.map-popup__commute--mid {
    --commute-accent: var(--v-theme-commute-mid);
}

.map-popup__commute--slow {
    --commute-accent: var(--v-theme-commute-slow);
}

.map-popup__commute-dot {
    flex: 0 0 auto;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background-color: rgb(var(--commute-accent));
    box-shadow: 0 0 0 3px color-mix(in srgb, rgb(var(--commute-accent)) 20%, transparent);
}

.map-popup__commute-text {
    font: inherit;
    line-height: 1.2;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
}

.map-popup__commute-text b {
    font-weight: 800;
    font-variant-numeric: tabular-nums;
    color: color-mix(in srgb, rgb(var(--commute-accent)) 80%, rgb(var(--v-theme-on-surface)));
}

.map-popup__actions {
    display: flex;
    gap: 10px;
    padding: 14px 18px 18px;
}

.map-popup__btn {
    flex: 1 1 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    height: 44px;
    border-radius: 999px;
    font-size: 14.5px;
    font-weight: 700;
    text-decoration: none;
    white-space: nowrap;
    transition:
        background-color 140ms ease,
        border-color 140ms ease,
        transform 60ms ease,
        box-shadow 140ms ease;
}

.map-popup__btn .mdi {
    font-size: 19px;
}

.map-popup__btn--primary {
    background-color: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-on-primary));
    box-shadow: 0 2px 12px -2px rgba(var(--v-theme-primary), 0.45);
}

.map-popup__btn--primary:hover {
    background-color: color-mix(in srgb, rgb(var(--v-theme-primary)) 88%, white);
    box-shadow: 0 5px 16px -3px rgba(var(--v-theme-primary), 0.55);
}

.map-popup__btn--ghost {
    flex: 0 0 auto;
    width: 48px;
    background: rgba(var(--v-theme-on-surface), 0.05);
    border: 1px solid rgba(var(--v-theme-on-surface), 0.12);
    color: rgb(var(--v-theme-on-surface));
}

.map-popup__btn--ghost:hover {
    background: rgba(var(--v-theme-on-surface), 0.09);
    border-color: rgba(var(--v-theme-on-surface), 0.22);
}

.map-popup__btn--ghost .mdi {
    color: inherit;
}

.map-popup__btn:active {
    transform: translateY(1px);
}

@media (max-width: 959.98px) {
    .leaflet-popup {
        --map-popup-width: min(292px, calc(100vw - 54px));
    }

    .map-popup__body {
        padding: 16px 14px 5px;
    }

    .map-popup__photo {
        height: 126px;
    }

    .map-popup__top {
        gap: 8px;
    }

    .map-popup__top-actions {
        top: 10px;
        right: 10px;
        gap: 6px;
    }

    .map-popup--no-photo .map-popup__header {
        padding-right: 72px;
    }

    .map-popup__fav,
    .map-popup__close {
        width: 30px;
        height: 30px;
    }

    .map-popup__price {
        font-size: 21px;
    }

    .map-popup__stats {
        margin: 14px 0 12px;
    }

    .map-popup__stat {
        padding: 10px 4px 9px;
    }

    .map-popup__stat-key {
        font-size: 9.5px;
    }

    .map-popup__commute-row {
        align-items: flex-start;
        gap: 8px;
    }

    .map-popup__commute {
        height: auto;
        min-height: 28px;
        max-width: 176px;
        padding-top: 5px;
        padding-bottom: 5px;
        white-space: normal;
    }

    .map-popup__actions {
        gap: 8px;
        padding: 12px 14px 16px;
    }

    .map-popup__btn {
        height: 42px;
    }

    .map-popup__btn--ghost {
        width: 44px;
    }
}
</style>
