import Vue from 'vue'
import Vuex from 'vuex'
import createLogger from 'vuex/logger'
import config from './modules/config'

Vue.use(Vuex)

const debug = process.env.NODE_ENV !== 'production'

export default new Vuex.Store({
  modules: {
    config
  },
  strict: debug,
  plugins: debug ? [createLogger()] : []
})
