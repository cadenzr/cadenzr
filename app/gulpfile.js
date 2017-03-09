var gulp = require("gulp");
var browserify = require("browserify");
var source = require('vinyl-source-stream');
var buffer = require('vinyl-buffer');
var uglify = require('gulp-uglify');
var ts = require('gulp-typescript');
var vueify = require('vueify')
var tsProject = ts.createProject('./src/tsconfig.json');
var copy = require('gulp-copy');
var sass = require('gulp-sass');


var workingDir = '.working';

gulp.task('sass', function () {
  return gulp.src('./assets/sass/**/*.scss')
    .pipe(sass().on('error', sass.logError))
    .pipe(gulp.dest('dist/assets/styles'));
});

gulp.task('copy-vue-components', function() {
	return gulp
		.src('src/**/*.vue')
		.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('copy-vendor', function() {
	return gulp
		.src('assets/vendor/*/**')
		.pipe(gulp.dest('dist/assets/vendor'));
});

gulp.task('typescript', function() {
	var tsResult = tsProject.src()
	.pipe(tsProject());

	return tsResult.js.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('browserify', ['copy-vue-components', 'typescript'], function() {
	    return browserify({
		            basedir: '.',
		            debug: true,
		            entries: [workingDir + '/js/main.js'],
		            cache: {},
		            packageCache: {}
		        })
		.transform(vueify)
	    .bundle()
	    .pipe(source('bundle.js'))
		.pipe(buffer())
		//.pipe(uglify())
	    .pipe(gulp.dest("dist/assets/scripts"));
});

gulp.task('watch-sass', function () {
	gulp.watch('assets/sass/**/*.scss' , ['sass']);
});


gulp.task('watch-ts', function () {
	gulp.watch('src/**/*.ts' , ['browserify']);
});


gulp.task('watch-vue', function () {
	gulp.watch('src/**/*.vue' , ['browserify']);
});

gulp.task('watch', ['watch-ts', 'watch-vue', 'watch-sass']);

gulp.task("default", ['copy-vendor', 'sass', 'browserify']);