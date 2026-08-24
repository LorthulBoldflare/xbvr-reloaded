<template>

</template>

<script>
import { Wampy } from 'wampy'

export default {
  name: 'Socket',
  data () {
    return {
      wsStatus: ''
    }
  },
  mounted () {
    const ws = new Wampy('/ws/', {
      realm: 'default',
      onConnect: () => {
        this.wsStatus = 'connected'
      },
      onClose: () => {
        this.wsStatus = 'disconnected'
      },
      onError: () => {
        this.wsStatus = 'disconnected'
      },
      onReconnect: () => {
        this.wsStatus = 'connecting'
      },
      onReconnectSuccess: () => {
        this.wsStatus = 'connected'
      }
    })    

    ws.subscribe('service.log', (dataArr, dataObj) => {
      if (dataArr.argsDict.level == 'debug') {
        console.debug(dataArr.argsDict.message)
      }
      if (dataArr.argsDict.level == 'info') {
        console.info(dataArr.argsDict.message)
      }
      if (dataArr.argsDict.level == 'error') {
        console.error(dataArr.argsDict.message)
      }

      if (dataArr.argsDict.data.task === 'scrape') {
        this.$store.commit('messages/setLastScrapeMessage', dataArr.argsDict)
      }

      if (dataArr.argsDict.data.task === 'scraperProgress') {
        if (dataArr.argsDict.message === 'DONE') {
          this.$store.commit('messages/setRunningScrapers', [])
        }

        if (dataArr.argsDict.data.started) {
          this.$store.commit('messages/addRunningScraper', dataArr.argsDict.data.scraperID)
        }

        if (dataArr.argsDict.data.completed) {
          this.$store.commit('messages/removeRunningScraper', dataArr.argsDict.data.scraperID)
        }
      }

      if (dataArr.argsDict.data.task === 'rescan') {
        this.$store.commit('messages/setLastRescanMessage', dataArr.argsDict)
      }
    })

    ws.subscribe('lock.change', (dataArr, dataObj) => {
      if (dataArr.argsDict.name === 'scrape') {
        this.$store.commit('messages/setLockScrape', dataArr.argsDict.locked)
      }
      if (dataArr.argsDict.name === 'rescan') {
        this.$store.commit('messages/setLockRescan', dataArr.argsDict.locked)
      }
    })

    ws.subscribe('state.change.optionsStorage', (arr, obj) => {
      this.$store.dispatch('optionsStorage/load')
    })

    ws.subscribe('options.previews.previewReady', (arr, obj) => {
      this.$store.commit('optionsPreviews/showPreview', { previewFn: arr.argsDict.previewFn })
    })

    ws.subscribe('options.previews.queue', (arr, obj) => {
      this.$store.commit('optionsPreviews/setQueue', arr.argsDict)
    })

    // Remote
    ws.subscribe('remote.state', (arr, obj) => {
      this.$store.dispatch('remote/processMessage', arr.argsDict)
    })
  }
}
</script>

<style scoped>

</style>
