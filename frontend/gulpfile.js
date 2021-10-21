const browserify = require('browserify')
const gulp = require('gulp')
const source = require('vinyl-source-stream')
const buffer = require('vinyl-buffer')
const hash = require('gulp-hash')
const uglify = require('gulp-uglify-es').default
const path = require('path')
const postcss = require('gulp-postcss')

function javascript () {
  const b = browserify({
    entries: './index.js',
    debug: true,
    transform: [
      [
        '@resonate/envlocalify', { NODE_ENV: 'development', global: true }
      ],
      ['babelify', {
        presets: ['@babel/preset-env'],
        plugins: [
          ['module:nanohtml'],
          ['@babel/plugin-transform-runtime', {
            absoluteRuntime: false,
            corejs: false,
            helpers: true,
            regenerator: true,
            useESModules: false
          }]
        ]
      }]
    ]
  })

  return b.bundle()
    .pipe(source('main.js'))
    .pipe(buffer())
    .pipe(uglify())
    .pipe(hash())
    .pipe(gulp.dest('../public/js'))
    .pipe(hash.manifest('../data/js/hash.json', {
      deleteOld: true,
      sourceDir: path.join(__dirname, '../public/js')
    }))
    .pipe(gulp.dest('.'))
}

function css () {
  return gulp.src('./index.css')
    .pipe(postcss([
      require('postcss-import')(),
      require('postcss-preset-env')({
        stage: 1,
        features: {
          browsers: ['last 1 version'],
          'nesting-rules': true
        }
      }),
      require('cssnano')({
        preset: ['default', {
          discardComments: {
            removeAll: true
          }
        }]
      })
    ]))
    .pipe(hash())
    .pipe(gulp.dest('../public/css'))
    .pipe(hash.manifest('../data/css/hash.json', {
      deleteOld: true,
      sourceDir: path.join(__dirname, '../public/css')
    }))
    .pipe(gulp.dest('.'))
}

gulp.task('javascript', javascript)
gulp.task('css', css)

gulp.task('watch:js', () => {
  gulp.watch('./src/**/*', javascript)
})

gulp.task('watch', () => {
  gulp.watch('./index.css', css)
  gulp.watch('./src/**/*', javascript)
})

gulp.task('default', gulp.series(javascript, css))
