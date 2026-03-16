<script lang="ts" setup>
import { computed } from 'vue'
import { Icon } from '@iconify/vue'

const props = defineProps<{
  position: string
  page: number
  size: number
  isPdf?: boolean
}>()

const emit = defineEmits<{
  'update:position': [value: string]
  'update:page': [value: number]
  'update:size': [value: number]
}>()

const positions = [
  { value: 'bottom-right', label: 'Kanan Bawah' },
  { value: 'bottom-left', label: 'Kiri Bawah' },
  { value: 'top-right', label: 'Kanan Atas' },
  { value: 'top-left', label: 'Kiri Atas' },
]

const pages = [
  { value: 0, label: 'Halaman Terakhir' },
  { value: 1, label: 'Halaman Pertama' },
]

const sizes = [
  { value: 20, label: 'Kecil (20mm)' },
  { value: 30, label: 'Sedang (30mm)' },
  { value: 40, label: 'Besar (40mm)' },
  { value: 50, label: 'Sangat Besar (50mm)' },
]

// Preview dimensions: A4 ratio = 210:297 ≈ 1:1.414
// Preview container width: 200px, height: 283px
const previewW = 200
const previewH = 283

// QR size relative to A4 width (210mm)
const qrPreviewSize = computed(() => {
  return Math.max(16, (props.size / 210) * previewW)
})

// Dynamic margin matching PDF: base 15pt + 10% of QR size (in points)
// Convert to preview pixels: 1pt ≈ 0.353mm, previewW/210 = px per mm
const margin = computed(() => {
  const sizePoints = props.size * 2.835
  const offsetPt = 15 + sizePoints / 10
  const offsetMM = offsetPt / 2.835
  return (offsetMM / 210) * previewW
})

// QR position in the preview
const qrStyle = computed(() => {
  const sz = qrPreviewSize.value
  const m = margin.value
  switch (props.position) {
    case 'top-left':
      return { top: `${m}px`, left: `${m}px`, width: `${sz}px`, height: `${sz}px` }
    case 'top-right':
      return { top: `${m}px`, right: `${m}px`, width: `${sz}px`, height: `${sz}px` }
    case 'bottom-left':
      return { bottom: `${m}px`, left: `${m}px`, width: `${sz}px`, height: `${sz}px` }
    case 'bottom-right':
    default:
      return { bottom: `${m}px`, right: `${m}px`, width: `${sz}px`, height: `${sz}px` }
  }
})

function selectPosition(pos: string) {
  emit('update:position', pos)
}
</script>

<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
    <h3 class="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
      <Icon icon="lucide:map-pin" class="w-4 h-4 text-emerald-600" />
      Posisi QR Code pada PDF
    </h3>
    <p class="text-xs text-gray-400 mb-3" v-if="!props.isPdf">Dokumen non-PDF: QR akan dihasilkan sebagai PDF terpisah</p>

    <div class="flex gap-5">
      <!-- Visual page preview -->
      <div class="flex-shrink-0">
        <p class="text-xs text-gray-500 mb-2 text-center">Preview</p>
        <div
          class="relative bg-white border-2 border-gray-300 rounded shadow-inner"
          :style="{ width: previewW + 'px', height: previewH + 'px' }"
        >
          <!-- Page lines (faux text lines) -->
          <div class="absolute inset-0 p-4 flex flex-col gap-1.5 pointer-events-none opacity-30">
            <div v-for="i in 14" :key="i" class="h-1.5 bg-gray-300 rounded-full"
              :style="{ width: (50 + Math.random() * 40) + '%' }" />
          </div>

          <!-- Clickable corner zones (small indicators that won't overlap with QR preview) -->
          <button
            v-for="pos in positions"
            :key="pos.value"
            @click="selectPosition(pos.value)"
            class="absolute z-20 w-5 h-5 flex items-center justify-center rounded-full transition-all duration-150"
            :class="position === pos.value
              ? 'bg-emerald-500 ring-2 ring-emerald-300 ring-offset-1'
              : 'bg-gray-300 hover:bg-gray-400'"
            :style="{
              top: pos.value.startsWith('top') ? '3px' : 'auto',
              bottom: pos.value.startsWith('bottom') ? '3px' : 'auto',
              left: pos.value.endsWith('left') ? '3px' : 'auto',
              right: pos.value.endsWith('right') ? '3px' : 'auto',
            }"
            :title="pos.label"
          >
            <div class="w-2 h-2 rounded-sm" :class="position === pos.value ? 'bg-white' : 'bg-white/60'" />
          </button>

          <!-- QR code preview square -->
          <div
            class="absolute bg-emerald-500/20 border-2 border-emerald-500 rounded-sm flex items-center justify-center transition-all duration-300 pointer-events-none"
            :style="qrStyle"
          >
            <Icon icon="lucide:qr-code" class="w-3/4 h-3/4 text-emerald-600" />
          </div>

          <!-- Page label -->
          <div class="absolute bottom-1 left-1/2 -translate-x-1/2 text-[9px] text-gray-400 pointer-events-none">
            {{ isPdf ? (page === 0 ? 'Hal. Terakhir' : 'Hal. Pertama') : 'PDF Baru' }}
          </div>
        </div>
      </div>

      <!-- Controls -->
      <div class="flex-1 space-y-3">
        <div>
          <label class="block text-xs text-gray-500 mb-1">Posisi</label>
          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="p in positions"
              :key="p.value"
              @click="selectPosition(p.value)"
              class="px-3 py-2 text-xs font-medium rounded-lg border transition-all"
              :class="position === p.value
                ? 'bg-emerald-50 border-emerald-500 text-emerald-700'
                : 'bg-white border-gray-200 text-gray-600 hover:border-gray-300'"
            >
              {{ p.label }}
            </button>
          </div>
        </div>

        <div v-if="props.isPdf">
          <label class="block text-xs text-gray-500 mb-1">Halaman</label>
          <select
            :value="page"
            @change="emit('update:page', Number(($event.target as HTMLSelectElement).value))"
            class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700 bg-white"
          >
            <option v-for="pg in pages" :key="pg.value" :value="pg.value">{{ pg.label }}</option>
          </select>
        </div>

        <div>
          <label class="block text-xs text-gray-500 mb-1">Ukuran QR Code</label>
          <div class="flex gap-2 flex-wrap">
            <button
              v-for="s in sizes"
              :key="s.value"
              @click="emit('update:size', s.value)"
              class="px-3 py-1.5 text-xs font-medium rounded-lg border transition-all"
              :class="size === s.value
                ? 'bg-emerald-50 border-emerald-500 text-emerald-700'
                : 'bg-white border-gray-200 text-gray-600 hover:border-gray-300'"
            >
              {{ s.label }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
