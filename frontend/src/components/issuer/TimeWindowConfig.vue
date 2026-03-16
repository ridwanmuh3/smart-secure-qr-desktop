<script lang="ts" setup>
  const validFrom = defineModel<string>('validFrom', { required: true })
  const validUntil = defineModel<string>('validUntil', { required: true })

  const emit = defineEmits<{
    preset: [value: string]
  }>()

  const presets = [
    { label: '1 Jam', value: 'hour' },
    { label: '24 Jam', value: 'day' },
    { label: '7 Hari', value: 'week' },
    { label: '30 Hari', value: 'month' },
  ]
</script>

<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
    <h3 class="text-sm font-semibold text-gray-700 mb-3 flex items-center gap-2">
      <span
        class="w-6 h-6 bg-emerald-100 text-emerald-700 rounded-full flex items-center justify-center text-xs font-bold">2</span>
      Jendela Waktu
    </h3>

    <div class="flex gap-2 mb-4">
      <button v-for="p in presets" :key="p.value" @click="emit('preset', p.value)"
        class="px-3 py-1.5 text-xs font-medium bg-emerald-50 text-emerald-700 hover:bg-emerald-100 border border-emerald-200 rounded-lg transition-colors cursor-pointer">
        {{ p.label }}
      </button>
    </div>

    <p class="text-xs text-gray-400 mb-3">Dokumen hanya dapat diverifikasi dalam rentang waktu yang ditentukan. Minimum
      30 detik dari sekarang.</p>

    <div class="grid grid-cols-2 gap-4">
      <div>
        <label class="block text-xs text-gray-500 mb-1">Berlaku Mulai (Waktu Buka)</label>
        <input v-model="validFrom" type="datetime-local"
          class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700" />
      </div>
      <div>
        <label class="block text-xs text-gray-500 mb-1">Berlaku Hingga</label>
        <input v-model="validUntil" type="datetime-local"
          class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700" />
      </div>
    </div>
  </div>
</template>
