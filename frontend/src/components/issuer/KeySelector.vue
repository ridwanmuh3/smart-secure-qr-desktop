<script lang="ts" setup>
import { computed } from 'vue'
import type { KeyPairInfo } from '../../types'

const modelValue = defineModel<string>({ required: true })

const props = defineProps<{
  keys: KeyPairInfo[]
}>()

const emit = defineEmits<{
  refresh: []
}>()

const signingKeys = computed(() => props.keys.filter(k => k.has_private_key))
</script>

<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-gray-700 flex items-center gap-2">
        <span class="w-6 h-6 bg-emerald-100 text-emerald-700 rounded-full flex items-center justify-center text-xs font-bold">3</span>
        Pilih Kunci Penandatangan
      </h3>
      <button @click="emit('refresh')" class="text-xs text-emerald-600 hover:text-emerald-700 font-medium">Muat ulang</button>
    </div>

    <div v-if="!keys.length" class="text-center py-4 text-gray-400 text-sm">
      Belum ada kunci. Buat kunci di halaman <router-link to="/keys" class="text-emerald-600 hover:underline">Kelola Kunci</router-link>.
    </div>

    <div v-else-if="!signingKeys.length" class="text-center py-4 text-amber-600 text-sm bg-amber-50 rounded-lg border border-amber-200 px-3">
      Tidak ada kunci dengan private key. Buat kunci baru atau impor private key di halaman <router-link to="/keys" class="font-medium hover:underline">Kelola Kunci</router-link>.
    </div>

    <select v-else v-model="modelValue"
      class="w-full px-3 py-2.5 text-sm border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700 bg-white">
      <option value="" disabled>Pilih kunci...</option>
      <option v-for="key in signingKeys" :key="key.id" :value="key.id">
        {{ key.name }} ({{ key.fingerprint }}) {{ key.is_default ? '⭐ Default' : '' }}
      </option>
    </select>
  </div>
</template>
