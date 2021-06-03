/* global XMLHttpRequest, FileReader, Image, Blob, FormData */

const Component = require('choo/component')

const html = require('choo/html')
const nanostate = require('nanostate')
const validateFormdata = require('validate-formdata')
const ProgressBar = require('../progress-bar')
const input = require('@resonate/input-element')

const uploadFile = (url, opts = {}, onProgress) => {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()

    xhr.upload.addEventListener('progress', onProgress)
    xhr.upload.addEventListener('loadend', () => {
      console.log('Ended')
    })

    xhr.open(opts.method || 'POST', url, true)
    xhr.withCredentials = true

    for (const k in opts.headers || {}) {
      xhr.setRequestHeader(k, opts.headers[k])
    }

    xhr.onload = e => {
      resolve(JSON.parse(e.target.response))
    }

    xhr.onerror = reject

    xhr.send(opts.body)
  })
}

const MAX_FILE_SIZE_IMAGE = 1024 * 1024 * 10

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

    this.machine = nanostate('idle', {
      idle: { drag: 'dragging', resolve: 'data' },
      dragging: { resolve: 'data', drag: 'idle' },
      data: { drag: 'dragging', resolve: 'data', reject: 'error' },
      error: { drag: 'idle', resolve: 'data' }
    })

    this.validator = validateFormdata()
    this.form = this.validator.state
  }

  onFileUploaded () {}

  createElement (props) {
    this.local.name = props.name || 'cover' // name ref for uploaded file
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

    const image = this.local.base64ImageData || this.local.src

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
      <div class="flex flex-${this.local.direction} ${this.machine.state === 'dragging' ? 'dragging' : ''}" unresolved>
        <div class="w-100">
          <div class="bg-image-placeholder flex relative" style="padding-top:calc(${props.format.height / props.format.width} * 100%);">
            <div style="background: url(${image}) center center / cover no-repeat;" class="upload absolute top-0 w-100 h-100 flex-auto">
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

        reader.onload = async e => {
          try {
            const base64FileData = reader.result.toString()

            this.local.base64ImageData = base64FileData

            const image = new Image()

            image.src = base64FileData
            image.onload = () => {
              this.width = image.width
              this.height = image.height
              this.validator.validate(`inputFile-${this._name}`, { width: this.width, height: this.height })
              this.rerender()
            }

            const formData = new FormData()
            formData.append('uploads', file)

            const response = await uploadFile('/api/user/upload', {
              method: 'POST',
              body: formData
            }, event => {
              if (event.lengthComputable) {
                const progress = event.loaded / event.total * 100
                this.local.progress = progress
                this.state.components[this._name + '-image-upload-progress'].slider.update({
                  value: this.local.progress
                })
              }
            })

            this.local.filename = response.data.filename

            this.validator.validate(this.local.name, this.local.filename)

            this.rerender()

            this.onFileUploaded(this.local.filename)
          } catch (err) {
            this.emit('error', err)
          }
        }

        reader.readAsDataURL(blob)
      }
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
    return props.src !== this.local.src
  }
}

module.exports = ImageUpload
