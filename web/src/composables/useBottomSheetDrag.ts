import { computed, onBeforeUnmount, ref, type CSSProperties } from "vue";

type DragOptions = {
    dismissThresholdPx?: number;
    dismissVelocityPxMs?: number;
    settleMs?: number;
};

const DEFAULT_DISMISS_THRESHOLD_PX = 96;
const DEFAULT_DISMISS_VELOCITY_PX_MS = 0.65;
const DEFAULT_SETTLE_MS = 210;

export function useBottomSheetDrag(
    onDismiss: () => void,
    options: DragOptions = {},
) {
    const offsetY = ref(0);
    const heightPx = ref<number | null>(null);
    const dragging = ref(false);
    const settling = ref(false);

    let pointerId: number | null = null;
    let startY = 0;
    let lastY = 0;
    let lastAt = 0;
    let velocityY = 0;
    let settleTimer: number | null = null;
    let resetTimer: number | null = null;
    let baseHeight = 0;

    const settleMs = options.settleMs ?? DEFAULT_SETTLE_MS;
    const dismissVelocity = options.dismissVelocityPxMs ?? DEFAULT_DISMISS_VELOCITY_PX_MS;

    const dragClasses = computed(() => ({
        "mobile-bottom-sheet--dragging": dragging.value,
        "mobile-bottom-sheet--settling": settling.value,
    }));

    const dragStyle = computed<CSSProperties>(() => {
        if (!dragging.value && !settling.value && offsetY.value === 0 && heightPx.value == null) {
            return {};
        }
        return {
            ...(offsetY.value !== 0 && { transform: `translate3d(0, ${offsetY.value}px, 0)` }),
            ...(heightPx.value != null && { height: `${heightPx.value}px` }),
        };
    });

    function clearSettleTimer(): void {
        if (settleTimer == null) return;
        window.clearTimeout(settleTimer);
        settleTimer = null;
    }

    function clearResetTimer(): void {
        if (resetTimer == null) return;
        window.clearTimeout(resetTimer);
        resetTimer = null;
    }

    function resetDrag(): void {
        clearSettleTimer();
        clearResetTimer();
        pointerId = null;
        offsetY.value = 0;
        heightPx.value = null;
        dragging.value = false;
        settling.value = false;
        velocityY = 0;
        baseHeight = 0;
    }

    function settleTo(y: number, afterSettle?: () => void): void {
        clearSettleTimer();
        clearResetTimer();
        dragging.value = false;
        settling.value = true;
        offsetY.value = y;
        if (baseHeight > 0) {
            heightPx.value = baseHeight;
        }
        settleTimer = window.setTimeout(() => {
            settleTimer = null;
            if (afterSettle) {
                afterSettle();
                resetTimer = window.setTimeout(() => {
                    resetTimer = null;
                    settling.value = false;
                    offsetY.value = 0;
                    heightPx.value = null;
                    baseHeight = 0;
                }, settleMs + 80);
                return;
            }
            settling.value = false;
            offsetY.value = 0;
            heightPx.value = null;
            baseHeight = 0;
        }, settleMs);
    }

    function onDragPointerDown(event: PointerEvent): void {
        if (event.button !== 0 || dragging.value || settling.value) return;
        const target = event.currentTarget;
        if (!(target instanceof HTMLElement)) return;

        clearSettleTimer();
        clearResetTimer();
        pointerId = event.pointerId;
        baseHeight = target.parentElement?.getBoundingClientRect().height ?? 0;
        startY = event.clientY;
        lastY = event.clientY;
        lastAt = performance.now();
        velocityY = 0;
        offsetY.value = 0;
        heightPx.value = null;
        dragging.value = true;
        settling.value = false;
        target.setPointerCapture(event.pointerId);
        event.preventDefault();
    }

    function onDragPointerMove(event: PointerEvent): void {
        if (!dragging.value || pointerId !== event.pointerId) return;

        const now = performance.now();
        const dt = Math.max(1, now - lastAt);
        const dy = event.clientY - startY;
        velocityY = (event.clientY - lastY) / dt;
        lastY = event.clientY;
        lastAt = now;
        if (dy < 0 && baseHeight > 0) {
            const viewportHeight =
                typeof window === "undefined" ? 800 : window.innerHeight;
            const maxHeight = Math.max(baseHeight, viewportHeight - 12);
            const extension = Math.min(-dy * 0.85, maxHeight - baseHeight);
            offsetY.value = 0;
            heightPx.value = baseHeight + extension;
        } else {
            offsetY.value = Math.max(0, dy);
            heightPx.value = baseHeight > 0 ? baseHeight : null;
        }
        event.preventDefault();
    }

    function finishDrag(event: PointerEvent): void {
        if (!dragging.value || pointerId !== event.pointerId) return;

        const target = event.currentTarget;
        if (target instanceof HTMLElement && target.hasPointerCapture(event.pointerId)) {
            target.releasePointerCapture(event.pointerId);
        }
        pointerId = null;

        const viewportHeight =
            typeof window === "undefined" ? 800 : window.innerHeight;
        const dismissThreshold = options.dismissThresholdPx ??
            Math.min(150, Math.max(DEFAULT_DISMISS_THRESHOLD_PX, viewportHeight * 0.16));
        const fastSwipe = offsetY.value > 28 && velocityY > dismissVelocity;
        const shouldDismiss = offsetY.value > dismissThreshold || fastSwipe;

        if (shouldDismiss) {
            settleTo(viewportHeight, onDismiss);
            return;
        }

        settleTo(0);
    }

    function onDragPointerCancel(event: PointerEvent): void {
        if (!dragging.value || pointerId !== event.pointerId) return;
        pointerId = null;
        settleTo(0);
    }

    onBeforeUnmount(resetDrag);

    return {
        dragClasses,
        dragStyle,
        resetDrag,
        onDragPointerDown,
        onDragPointerMove,
        onDragPointerUp: finishDrag,
        onDragPointerCancel,
    };
}
