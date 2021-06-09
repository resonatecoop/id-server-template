const browserify = require('browserify')
const gulp = require('gulp')
const source = require('vinyl-source-stream')
const buffer = require('vinyl-buffer')
const hash = require('gulp-hash')
const uglify = require('gulp-uglify-es').default
const references = require('gulp-hash-references')
const path = require('path')
const postcss = require('gulp-postcss')

function javascript () {
  const b = browserify({
    entries: './web/app/main.js',
    debug: true,
    transform: [
      [
        '@resonate/envlocalify', { NODE_ENV: 'development', global: true }
      ],
      ['babelify', {
        presets: ['@babel/preset-env'],
        plugins: [
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
    .pipe(gulp.dest('public/js'))
    .pipe(hash.manifest('data/js/hash.json', {
      deleteOld: true,
      sourceDir: path.join(__dirname, '/public/js')
    }))
    .pipe(gulp.dest('.'))
}

function css () {
  return gulp.src('./web/app/index.css')
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
    .pipe(gulp.dest('public/css'))
    .pipe(hash.manifest('data/css/hash.json', {
      deleteOld: true,
      sourceDir: path.join(__dirname, '/public/css')
    }))
    .pipe(gulp.dest('.'))
}

function derev () {
  return gulp.src('web/layouts/*.html')
    .pipe(references([
      path.join(__dirname, './data/js/hash.json'),
      path.join(__dirname, './data/css/hash.json')
    ], { dereference: true }))
    .pipe(gulp.dest('web/layouts'))
}

function rev () {
  return gulp.src('web/layouts/*.html')
    .pipe(references([
      path.join(__dirname, './data/js/hash.json'),
      path.join(__dirname, './data/css/hash.json')
    ]))
    .pipe(gulp.dest('web/layouts'))
}

gulp.task('javascript', gulp.series(derev, javascript, rev))
gulp.task('derev', derev)
gulp.task('rev', rev)
gulp.task('css', gulp.series(derev, css, rev))

gulp.task('watch', () => {
  gulp.watch('./web/app/index.css', css)
  gulp.watch('./web/app/**/*', javascript)
})
