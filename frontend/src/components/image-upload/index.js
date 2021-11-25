/* global XMLHttpRequest, fetch, FileReader, Image, Blob, FormData */

const Component = require('choo/component')

const html = require('choo/html')
const nanostate = require('nanostate')
const validateFormdata = require('validate-formdata')
const ProgressBar = require('../progress-bar')
const input = require('@resonate/input-element')
const imagePlaceholder = require('../../lib/image-placeholder')

/**
 * @function uploadFile
 * @description Upload file util function
 * @param {String} url Upload path (method POST by default)
 * @param {Object} opts xhr opts (method, headers, body)
 * @param {Function} onProgress optional onProgress callback function
 * @param {Function} onLoadEnd optional onLoadEnd callback function
 * @returns {Promise} Upload data response
 */
const uploadFile = (url = '/upload', opts = {}, onProgress = () => {}, loadend = () => {}) => {
  return new Promise((resolve, reject) => {
    const {
      headers = {},
      method = 'POST'
    } = opts

    const xhr = new XMLHttpRequest()

    xhr.upload.addEventListener('progress', onProgress)
    xhr.upload.addEventListener('loadend', loadend)

    xhr.open(method, url, true)
    xhr.withCredentials = true

    for (const k in headers) {
      xhr.setRequestHeader(k, headers[k])
    }

    xhr.onload = e => {
      resolve(JSON.parse(e.target.response))
    }

    xhr.onerror = reject

    xhr.send(opts.body)
  })
}

const uploadStatus = async (id, opts = {}) => {
  return await (await fetch(`/upload/${id}`, {
    method: 'GET',
    headers: {
      Pragma: 'no-cache',
      'Cache-Control': 'no-cache',
      Authorization: 'Bearer ' + opts.token
    }
  })).json()
}

const MAX_FILE_SIZE_IMAGE = 1024 * 1024 * 10

// ImageUpload component class
class ImageUpload extends Component {
  constructor (id, state, emit) {
    super(id)

    this.local = state.components[id] = {}
    this.state = state
    this.emit = emit

    this.local.progress = 0

    this.onDragOver = this.onDragOver.bind(this)
    this.onDragleave = this.onDragleave.bind(this)
    this.onChange = this.onChange.bind(this)
    this.getUploadStatus = this.getUploadStatus.bind(this)

    this.machine = nanostate('idle', {
      idle: { drag: 'dragging', resolve: 'data' },
      dragging: { resolve: 'data', drag: 'idle' },
      data: { drag: 'dragging', resolve: 'data', reject: 'error' },
      error: { drag: 'idle', resolve: 'data' }
    })

    this.local.checks = 0

    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  onFileUploaded () {}

  createElement (props) {
    this.local.name = props.name || 'cover' // name ref for uploaded file
    this.local.config = props.config || 'avatar'
    this.local.archive = props.archive

    this.validator = props.validator || this.validator
    this.form = props.form || this.form || {
      changed: false,
      valid: true,
      pristine: {},
      required: {},
      values: {},
      errors: {}
    }
    this.local.src = props.src

    this.onFileUploaded = props.onFileUploaded || this.onFileUploaded

    const errors = this.form.errors
    const values = this.form.values

    this.local.multiple = props.multiple || false
    this.local.format = props.format
    this.local.accept = props.accept || 'image/jpeg,image/jpg,image/png'
    this.local.direction = props.direction || 'row'
    this.local.ratio = props.ratio || '1200x1200px'

    const dropInfo = {
      idle: 'Drop an audio file',
      dragging: 'Drop now!',
      error: 'File not supported',
      data: 'Fetch Again?'
    }[this.machine.state]

    const image = this.local.objectURL || this.local.src || this.local.archive

    const fileInput = (options) => {
      const attrs = Object.assign({
        multiple: this.local.multiple,
        class: `w-100 h-100 o-0 absolute z-1 ${image ? 'loaded' : 'empty'}`,
        name: `inputFile-${this._name}`,
        required: false,
        onchange: this.onChange,
        title: dropInfo,
        accept: this.local.accept,
        type: 'file'
      }, options)

      return html`<input ${attrs}>`
    }

    return html`
      <div class="flex flex-column">
        <div class="flex flex-${this.local.direction} ${this.machine.state === 'dragging' ? 'dragging' : ''}" unresolved>
          <div class="w-100">
            <div class="bg-image-placeholder flex relative" style="padding-top:calc(${props.format.height / props.format.width} * 100%);">
              <div style="background: url(${image || imagePlaceholder(400, 400)}) center center / cover no-repeat;" class="absolute top-0 w-100 h-100 flex-auto z-1">
                <div class="relative w-100 h-100" ondragover=${this.onDragOver} ondrop=${this.onDrop} ondragleave=${this.onDragleave}>
                  ${fileInput({ id: `inputFile-${this._name}` })}
                  <label class="absolute o-0 w-100 h-100 top-0 left-0 right-0 bottom-0 z-1" style="cursor:pointer" for="inputFile-${this._name}">
                    Upload
                  </label>
                </div>
              </div>
            </div>
          </div>
          <div ondragover=${this.onDragOver} ondrop=${this.onDrop} ondragleave=${this.onDragleave} class="flex ${this.local.direction === 'row' ? 'ml3' : 'mt3'} flex-${this.local.direction === 'column' ? 'row' : 'column'}">
            <div class="relative grow mr2">
              ${fileInput({ id: `inputFile-${this._name}-button` })}
              <label class="dib pv2 ph4 mb1 ba bw b--black-80 ${this.direction === 'column' ? 'mr2' : ''}" for="inputFile-${this._name}-button">Upload</label>
            </div>
            ${errors[`inputFile-${this._name}`] || errors[`inputFile-${this._name}-button`]
              ? html`
                <p class="lh-copy f5 red">
                  ${errors[`inputFile-${this._name}`].message || errors[`inputFile-${this._name}-button`].message}
                </p>
              `
              : ''
            }
            ${errors[this.local.name] ? html`<p class="lh-copy f5 red">${errors[this.local.name].message}</p>` : ''}
            <div class="flex flex-column">
              <p class="lh-copy ma0 pa0 f6 grey">For best results, upload a JPG or PNG at ${this.local.ratio}</p>
              <div class="flex flex-column mt2">
                ${this.state.cache(ProgressBar, this._name + '-image-upload-progress').render({
                  progress: this.local.progress
                })}
              </div>
            </div>
            ${input({
              type: 'hidden',
              id: this.local.name,
              name: this.local.name,
              value: values[this.local.name]
            })}
          </div>
        </div>
      </div>
    `
  }

  onDragOver (e) {
    e.preventDefault()
    e.stopPropagation()
    if (this.machine.state === 'dragging') return false
    this.machine.emit('drag')

    this.rerender()
  }

  onDragleave (e) {
    e.preventDefault()
    e.stopPropagation()
    this.machine.emit('drag')
    this.rerender()
  }

  onDrop (e) {
  }

  onChange (e) {
    e.preventDefault()
    e.stopPropagation()

    this.local.uploading = true

    window.addEventListener('beforeunload', e => {
      if (!this.local.uploading) return

      e.preventDefault()
      e.returnValue = 'Upload in progress'
    })

    this.machine.emit('resolve')

    const files = e.target.files

    for (const file of files) {
      const reader = new FileReader()
      const size = file.size

      const image = ((/(image\/jpg|image\/jpeg|image\/png)/).test(file.type))

      if (!image) {
        this.machine.emit('reject')
        return this.rerender()
      }

      if (image) {
        if (size > MAX_FILE_SIZE_IMAGE) {
          this.machine.emit('reject')
          return this.rerender()
        }

        // Load some artwork
        const blob = new Blob([file], {
          type: file.type
        })

        this.local.objectURL = URL.createObjectURL(blob)

        reader.onload = async e => {
          try {
            const image = new Image()

            image.src = this.local.objectURL
            image.onload = () => {
              this.width = image.width
              this.height = image.height
              this.validator.validate(`inputFile-${this._name}`, { width: this.width, height: this.height })
              this.rerender()
            }

            const formData = new FormData()

            formData.append('uploads', file)
            formData.append('config', this.local.config)

            // upload file using upload tool (expect path /upload to be proxied to upload tool API)
            const response = await uploadFile('/upload', {
              method: 'POST',
              headers: {
                Authorization: 'Bearer ' + this.state.token
              },
              body: formData
            }, event => {
              if (event.lengthComputable) {
                // current progress by precentage
                const progress = event.loaded / event.total * 100

                this.local.progress = progress
                const componentID = this._name + '-image-upload-progress'
                // get slider component by reference
                const component = this.state.components[componentID]
                // update slider progress
                component.slider.update({
                  value: this.local.progress
                })
              }
            })

            this.local.filename = response.data.filename

            this.validator.validate(this.local.name, this.local.filename)

            this.rerender()

            this.getUploadStatus(this.local.filename)

            this.onFileUploaded(this.local.filename)
          } catch (err) {
            this.emit('error', err)
          }
        }

        reader.readAsDataURL(blob)
      }
    }
  }

  async getUploadStatus () {
    this.local.checks = this.local.checks + 1

    try {
      const response = await uploadStatus(this.local.filename, { token: this.state.token })

      if (response.status === 'ok') {
        this.local.uploading = false
        this.local.checks = 0
      } else if (this.local.checks <= 10) {
        setTimeout(() => {
          return this.getUploadStatus()
        }, 1000)
      }
    } catch (err) {
      this.emit('error', err)
    }
  }

  beforerender (el) {
    el.removeAttribute('unresolved')
  }

  afterupdate (el) {
    el.removeAttribute('unresolved')
  }

  load (el) {
    if (this.local.multiple) {
      const input = el.querySelector('input[type="file"]')
      input.attr('multiple', 'true')
    }
    this.validator.field(`inputFile-${this._name}`, { required: false }, (data) => {
      if (typeof data === 'object') {
        const { width, height } = data
        if (!width || !height) return new Error('Image is required')
        if (width < this.local.format.width || height < this.local.format.height) {
          return new Error('Image size is too small')
        }
      }
    })
    this.validator.field(`inputFile-${this._name}-button`, { required: false }, (data) => {
      if (typeof data === 'object') {
        const { width, height } = data
        if (!width || !height) return new Error('Image is required')
        if (width < this.local.format.width || height < this.local.format.height) {
          return new Error('Image size is too small')
        }
      }
    })
  }

  update (props) {
    return props.src !== this.local.src ||
      props.config !== this.local.config ||
      props.archive !== this.local.archive
  }
}

module.exports = ImageUpload
