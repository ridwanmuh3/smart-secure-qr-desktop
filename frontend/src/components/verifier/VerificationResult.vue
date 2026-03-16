<script lang="ts" setup>
  import { Icon } from '@iconify/vue'
  import type { VerificationResult } from '../../types'

  defineProps<{
    result: VerificationResult
  }>()

  const emit = defineEmits<{
    reset: []
  }>()

  const statusConfig: Record<string, { color: string; bg: string; border: string; icon: string; label: string }> = {
    authentic: {
      color: 'text-emerald-700', bg: 'bg-emerald-50', border: 'border-emerald-300',
      icon: 'lucide:circle-check',
      label: 'TERVERIFIKASI',
    },
    tampered: {
      color: 'text-red-700', bg: 'bg-red-50', border: 'border-red-300',
      icon: 'lucide:circle-x',
      label: 'TIDAK VALID',
    },
    not_yet_valid: {
      color: 'text-amber-700', bg: 'bg-amber-50', border: 'border-amber-300',
      icon: 'lucide:clock',
      label: 'BELUM BERLAKU',
    },
    expired: {
      color: 'text-gray-600', bg: 'bg-gray-100', border: 'border-gray-300',
      icon: 'lucide:triangle-alert',
      label: 'EXPIRED',
    },
    error: {
      color: 'text-red-700', bg: 'bg-red-50', border: 'border-red-300',
      icon: 'lucide:circle-alert',
      label: 'ERROR',
    },
  }

  function formatDateTime(dateStr: string): string {
    if (!dateStr) return '-'
    return new Date(dateStr).toLocaleString('id-ID', {
      year: 'numeric', month: 'long', day: 'numeric',
      hour: '2-digit', minute: '2-digit', second: '2-digit',
    })
  }

  function formatSize(bytes: number): string {
    if (!bytes) return '-'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }
</script>

<template>
  <div class="space-y-5">
    <!-- Status card -->
    <div :class="[statusConfig[result.status]?.bg, statusConfig[result.status]?.border]"
      class="rounded-xl border-2 p-6 text-center">
      <Icon :icon="statusConfig[result.status]?.icon" :class="statusConfig[result.status]?.color"
        class="w-16 h-16 mx-auto mb-3" />
      <h3 :class="statusConfig[result.status]?.color" class="text-2xl font-bold mb-2">
        {{ statusConfig[result.status]?.label }}
      </h3>
      <p class="text-gray-600">{{ result.message }}</p>
    </div>

    <!-- Anti-cloning warning -->
    <div v-if="result.scan_count > 5"
      class="rounded-xl border-2 border-amber-400 bg-amber-50 p-4 flex items-start gap-3">
      <Icon icon="lucide:triangle-alert" class="w-6 h-6 text-amber-600 shrink-0 mt-0.5" />
      <div>
        <p class="text-amber-800 font-semibold text-sm">Peringatan Anti-Kloning</p>
        <p class="text-amber-700 text-xs mt-1">
          QR code ini telah dipindai <strong>{{ result.scan_count }} kali</strong>.
          Jumlah pemindaian yang tinggi dapat mengindikasikan duplikasi atau kloning QR code.
        </p>
      </div>
    </div>

    <!-- Scan count info -->
    <div v-if="result.scan_count > 0 && result.scan_count <= 5"
      class="rounded-xl border border-gray-200 bg-white p-4 flex items-center gap-3">
      <Icon icon="lucide:eye" class="w-5 h-5 text-gray-400 shrink-0" />
      <p class="text-gray-600 text-sm">
        Total pemindaian: <strong>{{ result.scan_count }}</strong>
      </p>
    </div>

    <!-- Details (only for authentic) -->
    <div v-if="result.status === 'authentic'" class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-gray-700 mb-4">Detail Dokumen</h3>
      <div class="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
        <div>
          <span class="text-gray-500 text-xs">Nama File</span>
          <p class="text-gray-800 font-medium">{{ result.file_name }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Ukuran File</span>
          <p class="text-gray-800 font-medium">{{ formatSize(result.file_size) }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Penerbit</span>
          <p class="text-gray-800 font-medium">{{ result.issuer_id || '-' }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Ditandatangani Pada</span>
          <p class="text-gray-800 font-medium">{{ formatDateTime(result.issued_at) }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Berlaku Mulai</span>
          <p class="text-gray-800 font-medium">{{ formatDateTime(result.valid_from) }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Berlaku Hingga</span>
          <p class="text-gray-800 font-medium">{{ formatDateTime(result.valid_until) }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Total Pemindaian</span>
          <p class="text-gray-800 font-medium">{{ result.scan_count }}</p>
        </div>
        <div class="col-span-2">
          <span class="text-gray-500 text-xs">Hash Dokumen</span>
          <p class="text-gray-800 font-mono text-xs break-all">{{ result.document_hash }}</p>
        </div>
        <div class="col-span-2">
          <span class="text-gray-500 text-xs">Fingerprint Kunci Publik</span>
          <p class="text-gray-800 font-mono text-xs">{{ result.public_key_hex }}</p>
        </div>
        <div v-if="result.metadata" class="col-span-2">
          <span class="text-gray-500 text-xs">Metadata</span>
          <p class="text-gray-800">{{ result.metadata }}</p>
        </div>
      </div>
    </div>

    <!-- Time info for not_yet_valid / expired -->
    <div v-if="result.status === 'not_yet_valid' || result.status === 'expired'"
      class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
      <h3 class="text-sm font-semibold text-gray-700 mb-3">Informasi Waktu</h3>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-gray-500 text-xs">Berlaku Mulai</span>
          <p class="text-gray-800 font-medium">{{ formatDateTime(result.valid_from) }}</p>
        </div>
        <div>
          <span class="text-gray-500 text-xs">Berlaku Hingga</span>
          <p class="text-gray-800 font-medium">{{ formatDateTime(result.valid_until) }}</p>
        </div>
      </div>
    </div>

    <button @click="emit('reset')"
      class="w-full py-2.5 bg-gray-100 hover:bg-gray-200 text-gray-700 font-medium rounded-lg transition-colors cursor-pointer">
      Verifikasi QR Lainnya
    </button>
  </div>
</template>
