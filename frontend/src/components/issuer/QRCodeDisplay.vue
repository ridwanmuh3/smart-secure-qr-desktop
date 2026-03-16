<script lang="ts" setup>
import { Icon } from '@iconify/vue'

defineProps<{
  qrBase64: string
  secureId: string
  documentHash: string
  fileName: string
  validFrom: string
  validUntil: string
  signedFilePath?: string
  isPdf?: boolean
}>()

const emit = defineEmits<{
  save: []
  saveSigned: []
  reset: []
}>()

function formatDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('id-ID', {
    year: 'numeric', month: 'long', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  })
}
</script>

<template>
  <div class="space-y-5">
    <div class="bg-white rounded-xl border border-emerald-200 p-6 shadow-sm text-center">
      <div class="inline-flex items-center gap-2 px-3 py-1.5 bg-emerald-100 text-emerald-700 rounded-full text-sm font-medium mb-4">
        <Icon icon="lucide:check" class="w-4 h-4" />
        {{ isPdf ? 'QR Code Disisipkan ke PDF' : 'QR Code & PDF Berhasil Dibuat' }}
      </div>

      <div class="flex justify-center mb-4">
        <img :src="'data:image/png;base64,' + qrBase64" alt="Secure QR Code"
          class="w-64 h-64 rounded-lg shadow-md border border-gray-100" />
      </div>

      <div class="flex justify-center gap-3 mb-6">
        <button v-if="signedFilePath" @click="emit('saveSigned')"
          class="px-5 py-2.5 bg-emerald-600 hover:bg-emerald-700 text-white font-medium rounded-lg transition-colors flex items-center gap-2 cursor-pointer">
          <Icon icon="lucide:file-down" class="w-4 h-4" />
          {{ isPdf ? 'Simpan PDF Tertandatangani' : 'Simpan PDF QR Code' }}
        </button>
        <button @click="emit('save')"
          class="px-5 py-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium rounded-lg transition-colors flex items-center gap-2">
          <Icon icon="lucide:download" class="w-4 h-4" />
          Simpan Gambar QR
        </button>
        <button @click="emit('reset')"
          class="px-5 py-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium rounded-lg transition-colors">
          Buat Baru
        </button>
      </div>
    </div>

    <!-- Summary info -->
    <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-gray-700 mb-3">Detail QR Code</h3>
      <div class="space-y-2 text-sm">
        <div class="flex justify-between">
          <span class="text-gray-500">Secure ID</span>
          <span class="font-mono text-gray-700 text-xs">{{ secureId }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-500">Dokumen</span>
          <span class="text-gray-700">{{ fileName }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-500">Hash Dokumen</span>
          <span class="font-mono text-gray-700 text-xs truncate max-w-xs">{{ documentHash }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-500">Berlaku Mulai</span>
          <span class="text-gray-700">{{ formatDateTime(validFrom) }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-500">Berlaku Hingga</span>
          <span class="text-gray-700">{{ formatDateTime(validUntil) }}</span>
        </div>
        <div v-if="signedFilePath" class="flex justify-between">
          <span class="text-gray-500">Output PDF</span>
          <span class="text-emerald-600 text-xs font-medium">{{ isPdf ? 'QR disisipkan dalam dokumen' : 'PDF sertifikat QR' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
