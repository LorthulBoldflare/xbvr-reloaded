import api from '../api'

// confirmAndDeleteAkaGroup shows the standard confirmation dialog and deletes
// the aka group for the given actor name. Shared by the scenes filters,
// actors filters and actor details views. onDone runs after a successful
// delete (clear local state, reload the list).
export function confirmAndDeleteAkaGroup (vm, name, storeModule, onDone) {
  vm.$buefy.dialog.confirm({
    title: vm.$t('Delete aka group'),
    message: `Do you want to delete the aka group for <strong>${name}</strong>?`,
    type: 'is-danger',
    hasIcon: true,
    confirmText: vm.$t('Delete'),
    onConfirm: () => {
      vm.$store.state[storeModule].isLoading = true
      api.post('/aka/delete', { json: { name } }).json().then(() => {
        onDone()
        vm.$store.dispatch(storeModule + '/filters')
        vm.$store.state[storeModule].isLoading = false
      })
    }
  })
}
