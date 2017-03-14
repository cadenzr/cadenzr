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
var livereload = require('gulp-livereload');
var htmlreplace = require('gulp-html-replace');
var util = require('gulp-util');

var workingDir = '.working';

gulp.task('index', function () {
	var bundle = '/assets/scripts/bundle.';
	if(util.env.production) {
		bundle += 'min.';
	}
	bundle += 'js';

	return gulp.src('index.html')
		.pipe(htmlreplace({
		'bundle': bundle
		}))
		.pipe(gulp.dest('dist/'))
		.pipe(livereload());
});

gulp.task('sass', function () {
  return gulp.src('./assets/sass/**/*.scss')
    .pipe(sass().on('error', sass.logError))
    .pipe(gulp.dest('dist/assets/styles'))
	.pipe(livereload());
});

gulp.task('copy-vue-components', function() {
	return gulp
		.src('src/**/*.vue')
		.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('copy-vendor', function() {
	return gulp
		.src('assets/vendor/**/*')
		.pipe(gulp.dest('dist/assets/vendor'));
});

gulp.task('copy-images', function() {
    return gulp
        .src('assets/images/**/*')
        .pipe(gulp.dest('dist/assets/images'));
});

gulp.task('typescript', ['copy-vue-components'], function() {
	var tsResult = tsProject.src()
	.pipe(tsProject());

	return tsResult.js.pipe(gulp.dest(workingDir + '/js'));
});

gulp.task('browserify', ['typescript'], function() {
		var bundle = 'bundle.';
		if(util.env.production) {
			bundle += 'min.';
		}
		bundle += 'js';

	    var r = browserify({
		            basedir: '.',
		            debug: true,
		            entries: [workingDir + '/js/main.js'],
		            cache: {},
		            packageCache: {}
		        })
		.transform(vueify)
	    .bundle()
	    .pipe(source(bundle))
		.pipe(buffer());

		if(util.env.production) {
			r.pipe(uglify())
		}

	    r.pipe(gulp.dest("dist/assets/scripts"))
		r.pipe(livereload());

		return r;
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

gulp.task('watch-index', function () {
	gulp.watch('index.html' , ['index']);
});

gulp.task('watch', ['watch-index', 'watch-ts', 'watch-vue', 'watch-sass'], function() {
	livereload.listen();
});

gulp.task("default", ['index', 'copy-vendor', 'copy-images', 'sass', 'browserify']);