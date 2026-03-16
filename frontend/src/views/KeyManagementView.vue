<script lang="ts" setup>
  import { onMounted, ref } from 'vue'
  import { Icon } from '@iconify/vue'
  import { useKeyManagement } from '../composables/useKeyManagement'

  const {
    keys, isLoading, error,
    loadKeys, generateKey, deleteKey, setDefault,
    exportKey, exportPrivate, importKey, importPrivate,
  } = useKeyManagement()

  const showGenDialog = ref(false)
  const newKeyName = ref('')
  const showImportDialog = ref(false)
  const importKeyName = ref('')
  const importType = ref<'public' | 'secret'>('public')
  const showDeleteConfirm = ref<string | null>(null)
  const deleteTargetName = ref('')

  onMounted(() => loadKeys())

  async function handleGenerate() {
    if (!newKeyName.value.trim()) return
    await generateKey(newKeyName.value.trim())
    newKeyName.value = ''
    showGenDialog.value = false
  }

  async function handleImport() {
    if (!importKeyName.value.trim()) return
    if (importType.value === 'secret') {
      await importPrivate(importKeyName.value.trim())
    } else {
      await importKey(importKeyName.value.trim())
    }
    importKeyName.value = ''
    importType.value = 'public'
    showImportDialog.value = false
  }

  function confirmDelete(id: string, name: string) {
    showDeleteConfirm.value = id
    deleteTargetName.value = name
  }

  async function handleDelete() {
    if (!showDeleteConfirm.value) return
    await deleteKey(showDeleteConfirm.value)
    showDeleteConfirm.value = null
  }

  function formatDate(dateStr: string) {
    return new Date(dateStr).toLocaleDateString('id-ID', {
      year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit',
    })
  }
</script>

<template>
  <div class="p-6 max-w-4xl mx-auto">
    <div class="flex items-center justify-between mb-6">
      <div>
        <h2 class="text-2xl font-bold text-gray-800">Kelola Kunci</h2>
        <p class="text-gray-500 text-sm mt-1">Buat, impor, dan kelola pasangan kunci digital</p>
      </div>
      <div class="flex gap-2">
        <button @click="showImportDialog = true"
          class="px-4 py-2 text-sm font-medium text-emerald-700 bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 rounded-lg transition-colors cursor-pointer">
          Impor Kunci
        </button>
        <button @click="showGenDialog = true"
          class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 rounded-lg transition-colors flex items-center gap-1.5 cursor-pointer">
          <Icon icon="lucide:plus" class="w-4 h-4" />
          Buat Kunci Baru
        </button>
      </div>
    </div>

    <!-- Error alert -->
    <div v-if="error" class="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
      {{ error }}
    </div>

    <!-- Loading -->
    <div v-if="isLoading" class="text-center py-12 text-gray-500">Memuat kunci...</div>

    <!-- Empty state -->
    <div v-else-if="!keys.length" class="text-center py-16 bg-white rounded-xl border border-gray-200">
      <Icon icon="lucide:key-round" class="w-16 h-16 text-gray-300 mx-auto mb-4" />
      <p class="text-gray-500 mb-2">Belum ada kunci</p>
      <p class="text-gray-400 text-sm">Buat pasangan kunci baru untuk mulai menandatangani dokumen</p>
    </div>

    <!-- Key list -->
    <div v-else class="space-y-3">
      <div v-for="key in keys" :key="key.id"
        class="bg-white rounded-xl border p-4 shadow-sm transition-all hover:shadow-md"
        :class="key.is_default ? 'border-emerald-300 ring-1 ring-emerald-200' : 'border-gray-200'">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-lg flex items-center justify-center" :class="key.has_private_key
              ? (key.is_default ? 'bg-emerald-100 text-emerald-600' : 'bg-blue-100 text-blue-600')
              : 'bg-gray-100 text-gray-400'">
              <Icon icon="lucide:key-round" class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="font-semibold text-gray-800">{{ key.name }}</span>
                <span v-if="key.is_default"
                  class="px-2 py-0.5 text-xs font-medium bg-emerald-100 text-emerald-700 rounded-full">Default</span>
                <span v-if="key.has_private_key"
                  class="px-2 py-0.5 text-xs font-medium bg-blue-100 text-blue-700 rounded-full">
                  Kunci Lengkap
                </span>
                <span v-else class="px-2 py-0.5 text-xs font-medium bg-gray-100 text-gray-500 rounded-full">
                  Kunci Publik
                </span>
              </div>
              <div class="text-xs text-gray-400 mt-0.5 font-mono">{{ key.fingerprint }}</div>
            </div>
          </div>
          <div class="flex items-center gap-1.5">
            <span class="text-xs text-gray-400 mr-2">{{ formatDate(key.created_at) }}</span>
            <button v-if="!key.is_default" @click="setDefault(key.id)"
              class="px-2.5 py-1.5 text-xs text-gray-500 hover:text-emerald-600 hover:bg-emerald-50 rounded-lg transition-colors flex items-center gap-1 cursor-pointer">
              <Icon icon="lucide:star" class="w-3.5 h-3.5" />
              <span>Default</span>
            </button>
            <button @click="exportKey(key.id)"
              class="px-2.5 py-1.5 text-xs text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors flex items-center gap-1 cursor-pointer">
              <Icon icon="lucide:download" class="w-3.5 h-3.5" />
              <span>Publik</span>
            </button>
            <button v-if="key.has_private_key" @click="exportPrivate(key.id)"
              class="px-2.5 py-1.5 text-xs text-gray-500 hover:text-amber-600 hover:bg-amber-50 rounded-lg transition-colors flex items-center gap-1 cursor-pointer">
              <Icon icon="lucide:lock" class="w-3.5 h-3.5" />
              <span>Privat</span>
            </button>
            <button @click="confirmDelete(key.id, key.name)"
              class="px-2.5 py-1.5 text-xs text-gray-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors flex items-center gap-1 cursor-pointer">
              <Icon icon="lucide:trash-2" class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Generate Key Dialog -->
    <Teleport to="body">
      <div v-if="showGenDialog" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50"
        @click.self="showGenDialog = false">
        <div class="bg-white rounded-2xl p-6 w-96 shadow-2xl">
          <h3 class="text-lg font-bold text-gray-800 mb-4">Buat Pasangan Kunci Baru</h3>
          <div class="mb-4">
            <label class="block text-sm text-gray-600 mb-1">Nama Kunci</label>
            <input v-model="newKeyName" type="text" placeholder="Contoh: Kunci Utama"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700"
              @keyup.enter="handleGenerate" autofocus />
          </div>
          <p class="text-xs text-gray-400 mb-4">Akan membuat pasangan kunci baru (public + private key)</p>
          <div class="flex gap-2 justify-end">
            <button @click="showGenDialog = false"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors cursor-pointer">Batal</button>
            <button @click="handleGenerate" :disabled="!newKeyName.trim()"
              class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 disabled:bg-gray-300 rounded-lg transition-colors cursor-pointer">Buat</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Import Key Dialog -->
    <Teleport to="body">
      <div v-if="showImportDialog" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50"
        @click.self="showImportDialog = false">
        <div class="bg-white rounded-2xl p-6 w-96 shadow-2xl">
          <h3 class="text-lg font-bold text-gray-800 mb-4">Impor Kunci</h3>
          <div class="mb-4">
            <label class="block text-sm text-gray-600 mb-1">Nama Kunci</label>
            <input v-model="importKeyName" type="text" placeholder="Contoh: Kunci Verifikasi"
              class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:ring-2 focus:ring-emerald-500 focus:border-emerald-500 outline-none text-gray-700"
              @keyup.enter="handleImport" autofocus />
          </div>
          <div class="mb-4">
            <label class="block text-sm text-gray-600 mb-2">Tipe Kunci</label>
            <div class="flex gap-3">
              <label class="flex items-center gap-2 px-3 py-2 border rounded-lg cursor-pointer transition-colors"
                :class="importType === 'public' ? 'border-emerald-300 bg-emerald-50' : 'border-gray-200 hover:bg-gray-50'">
                <input type="radio" v-model="importType" value="public" class="text-emerald-600" />
                <div>
                  <span class="text-sm font-medium text-gray-700">Kunci Publik</span>
                  <p class="text-xs text-gray-400">Hanya verifikasi</p>
                </div>
              </label>
              <label class="flex items-center gap-2 px-3 py-2 border rounded-lg cursor-pointer transition-colors"
                :class="importType === 'secret' ? 'border-amber-300 bg-amber-50' : 'border-gray-200 hover:bg-gray-50'">
                <input type="radio" v-model="importType" value="secret" class="text-amber-600" />
                <div>
                  <span class="text-sm font-medium text-gray-700">Kunci Privat</span>
                  <p class="text-xs text-gray-400">Tanda tangan + verifikasi</p>
                </div>
              </label>
            </div>
          </div>
          <p class="text-xs text-gray-400 mb-4">
            {{ importType === 'secret' ? 'Pilih file private key untuk menandatangani dan memverifikasi' : 'Pilih file kunci publik yang akan diimpor(hanya untuk verifikasi' }}
          </p>
          <div class="flex gap-2 justify-end">
            <button @click="showImportDialog = false; importType = 'public'"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors cursor-pointer">Batal</button>
            <button @click="handleImport" :disabled="!importKeyName.trim()"
              class="px-4 py-2 text-sm font-medium text-white bg-emerald-600 hover:bg-emerald-700 disabled:bg-gray-300 rounded-lg transition-colors cursor-pointer">Impor</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirmation Dialog -->
    <Teleport to="body">
      <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/40 flex items-center justify-center z-50"
        @click.self="showDeleteConfirm = null">
        <div class="bg-white rounded-2xl p-6 w-96 shadow-2xl">
          <div class="flex items-center gap-3 mb-4">
            <div class="w-10 h-10 bg-red-100 rounded-full flex items-center justify-center">
              <Icon icon="lucide:triangle-alert" class="w-5 h-5 text-red-600" />
            </div>
            <h3 class="text-lg font-bold text-gray-800">Hapus Kunci</h3>
          </div>
          <p class="text-sm text-gray-600 mb-1">Apakah Anda yakin ingin menghapus kunci <strong>{{ deleteTargetName
              }}</strong>?</p>
          <p class="text-xs text-red-500 mb-4">Tindakan ini tidak dapat dibatalkan. Dokumen yang ditandatangani dengan
            kunci ini tidak akan dapat diverifikasi.</p>
          <div class="flex gap-2 justify-end">
            <button @click="showDeleteConfirm = null"
              class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-lg transition-colors cursor-pointer">Batal</button>
            <button @click="handleDelete"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors cursor-pointer">Hapus</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
