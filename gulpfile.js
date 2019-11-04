const browserify = require('browserify')
const gulp = require('gulp')
const source = require('vinyl-source-stream')
const buffer = require('vinyl-buffer')
const uglify = require('gulp-uglify-es').default
const del = require('del')
const postcss = require('gulp-postcss')

function javascript () {
  const b = browserify({
    entries: './web/js/main.js',
    debug: true,
    transform: [
      ['babelify', { presets: ['@babel/preset-env'] }]
    ]
  })

  del(['public/js/**/*'])

  return b.bundle()
    .pipe(source('main.js'))
    .pipe(buffer())
    .pipe(uglify())
    .pipe(gulp.dest('public/js'))
    .pipe(gulp.dest('data/js'))
}

function css () {
  del(['public/css/**/*'])

  return gulp.src('./web/css/main.css')
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
    .pipe(gulp.dest('static/dist/css'))
    .pipe(gulp.dest('data/css'))
}

gulp.task('javascript', javascript)
gulp.task('css', css)

gulp.task('watch', () => {
  gulp.watch('./css/**/*', css)
  gulp.watch('./js/**/*', javascript)
})

gulp.task('default', gulp.series(gulp.parallel('javascript', 'css')))
