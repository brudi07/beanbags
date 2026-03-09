<script setup lang="ts">
import { ref, onMounted, onUnmounted } from "vue"
import { v4 as uuidv4 } from "uuid"
import { useScoringStore } from "~/stores/scoringStore"
import type { ThrowData } from "~/types/game"

const scoringStore = useScoringStore()

const boardRef = ref<HTMLElement | null>(null)
const missZoneRef = ref<HTMLElement | null>(null)
const itoZoneRef = ref<HTMLElement | null>(null)

let draggingBagId: string | null = null

// Hole settings
const HOLE_X = 0.5
const HOLE_Y = 0.3
const HOLE_RADIUS = 0.045
const HOLE_MAGNET_RADIUS = 0.07

function detectResult(x: number, y: number) {
    const dx = x - HOLE_X
    const dy = y - HOLE_Y
    const distance = Math.sqrt(dx * dx + dy * dy)

    if (distance < HOLE_RADIUS) return "hole"

    if (x >= 0 && x <= 1 && y >= 0 && y <= 1) return "board"

    if (y > 1 && y <= 1.17) return "miss"

    if (y > 1.17) return "ito"

    return "miss"
}

function startNewBag(team: 1 | 2, event: PointerEvent) {

    const id = uuidv4()

    const newBag: ThrowData & { rotation?: number } = {
        id,
        team,
        playerId: "player",
        x: 0.5,
        y: 1.1,
        result: "miss",
        round: scoringStore.round,
        timestamp: Date.now(),
        rotation: Math.random() * 20 - 10
    }

    scoringStore.addThrow(newBag)

    if (team === 1) scoringStore.team1BagsRemaining--
    else scoringStore.team2BagsRemaining--

    draggingBagId = id

    moveBag(event)
}

function startDrag(id: string, event: PointerEvent) {
    draggingBagId = id
    moveBag(event)
}

function moveBag(event: PointerEvent) {

    if (!draggingBagId) return

    const bag = scoringStore.throws.find(t => t.id === draggingBagId)
    if (!bag) return

    const boardRect = boardRef.value?.getBoundingClientRect()
    const missRect = missZoneRef.value?.getBoundingClientRect()
    const itoRect = itoZoneRef.value?.getBoundingClientRect()

    if (!boardRect) return

    // Check zones FIRST - before calculating board-relative position

    if (
        missRect &&
        event.clientX >= missRect.left &&
        event.clientX <= missRect.right &&
        event.clientY >= missRect.top &&
        event.clientY <= missRect.bottom
    ) {
        bag.result = "miss"
        // Position relative to miss zone - spread bags horizontally
        const relativeX = (event.clientX - missRect.left) / missRect.width
        bag.x = relativeX * 0.45 + 0.05  // Maps to left side of board area (0.05 to 0.5)
        bag.y = 1.08
        return
    }

    if (
        itoRect &&
        event.clientX >= itoRect.left &&
        event.clientX <= itoRect.right &&
        event.clientY >= itoRect.top &&
        event.clientY <= itoRect.bottom
    ) {
        bag.result = "ito"
        // Position relative to ITO zone - spread bags horizontally
        const relativeX = (event.clientX - itoRect.left) / itoRect.width
        bag.x = relativeX * 0.45 + 0.5  // Maps to right side of board area (0.5 to 0.95)
        bag.y = 1.08
        return
    }

    // Calculate position relative to board
    let x = (event.clientX - boardRect.left) / boardRect.width
    let y = (event.clientY - boardRect.top) / boardRect.height

    // allow dragging above and below board, but clamp x to board bounds
    x = Math.max(0, Math.min(1, x))
    y = Math.max(-0.2, Math.min(1.4, y))

    // hole magnet assist
    const dx = x - HOLE_X
    const dy = y - HOLE_Y
    const distance = Math.sqrt(dx * dx + dy * dy)

    if (distance < HOLE_MAGNET_RADIUS) {
        bag.x = HOLE_X
        bag.y = HOLE_Y
    } else {
        bag.x = x
        bag.y = y
    }

    bag.result = detectResult(bag.x, bag.y)
}

function stopDrag() {
    draggingBagId = null
}

onMounted(() => {
    window.addEventListener("pointermove", moveBag)
    window.addEventListener("pointerup", stopDrag)
})

onUnmounted(() => {
    window.removeEventListener("pointermove", moveBag)
    window.removeEventListener("pointerup", stopDrag)
})
</script>

<template>

    <div class="flex flex-col items-center gap-6 w-full">

        <!-- BOARD -->

        <div ref="boardRef"
            class="relative w-full max-w-md aspect-[2/3] bg-gradient-to-b from-amber-200 to-amber-400 rounded-xl border shadow-lg touch-none">

            <!-- HOLE -->

            <div class="absolute rounded-full border-4 border-gray-700 bg-black shadow-inner" :style="{
                width: '90px',
                height: '90px',
                left: '50%',
                top: '30%',
                transform: 'translate(-50%, -50%)'
            }" />

            <!-- BAGS -->

            <div v-for="(bag, index) in scoringStore.throws" :key="bag.id"
                class="absolute w-14 h-14 rounded-md cursor-pointer touch-none select-none shadow-xl transition-transform"
                :class="[
                    bag.team === 1 ? 'bg-red-500' : 'bg-blue-500',
                    bag.result === 'miss' ? 'opacity-40' : '',
                    bag.result === 'ito' ? 'opacity-40 border-2 border-yellow-500' : '',
                    bag.result === 'hole' ? 'scale-75 opacity-70' : ''
                ]" :style="{
                    left: bag.x * 100 + '%',
                    top: bag.y * 100 + '%',
                    transform: `translate(-50%, -50%) rotate(${bag.rotation || 0}deg)`,
                    zIndex: index
                }" @pointerdown="startDrag(bag.id, $event)" />

        </div>

        <!-- MISS / ITO ZONES -->

        <div class="w-full max-w-md grid grid-cols-2 gap-4">

            <div ref="missZoneRef"
                class="h-20 border-2 border-dashed border-gray-400 rounded-lg flex items-center justify-center text-sm text-gray-500 bg-gray-100">
                Missed Throw
            </div>

            <div ref="itoZoneRef"
                class="h-20 border-2 border-dashed border-yellow-500 rounded-lg flex items-center justify-center text-sm text-yellow-700 bg-yellow-100">
                Intentional Throw Off
            </div>

        </div>

        <!-- BAG POOLS -->

        <div class="flex justify-between w-full max-w-md">

            <!-- TEAM 1 -->

            <div class="flex flex-col items-center gap-2">
                <p class="font-semibold text-sm">Team 1</p>

                <div class="flex gap-3 mr-1">

                    <div v-for="n in scoringStore.team1BagsRemaining" :key="'t1-' + n"
                        class="w-12 h-12 bg-red-500 rounded-md cursor-pointer shadow"
                        @pointerdown="startNewBag(1, $event)" />

                </div>
            </div>

            <!-- TEAM 2 -->

            <div class="flex flex-col items-center gap-2">
                <p class="font-semibold text-sm">Team 2</p>

                <div class="flex gap-3 ml-1">

                    <div v-for="n in scoringStore.team2BagsRemaining" :key="'t2-' + n"
                        class="w-12 h-12 bg-blue-500 rounded-md cursor-pointer shadow"
                        @pointerdown="startNewBag(2, $event)" />

                </div>
            </div>

        </div>

    </div>

</template>