#include<stdio.h>
#include<stdlib.h>
// アライメントはなしで
#pragma pack(1)
typedef struct {
  unsigned short byType;
  unsigned int bfSize;
  unsigned short bfReserved1;
  unsigned short bfReserved2;
  unsigned int bfOffBits;
} file_header;
#pragma pack(0)

#pragma pack(1)
typedef struct {
  unsigned int biSize;
  int biWidth;
  int biHeight;
  unsigned short biPlanes;
  unsigned short biBitCount;
  unsigned int biCompression;
  unsigned int biSizeImage;
  int biXPixPerMeter;
  int BiYPixPerMeter;
  unsigned int biClrUserd;
  unsigned int biCirImportant;
} information_header;
#pragma pack(0)

int blackBitsNum(char byte) {
  int num = 0;
  int i;
  for (i=0; i < 8; i++) {
    // byte の下位1bit が0である byte & 0x0 ではだめ
    if ( !(byte & 0x1) ) {
      num++;
    }
    byte = byte >> 1;
  }
  return num;
}

unsigned char* divide(unsigned char* pixel, int width, int height, int divide)
{
  int line = width / 2;
  int l,h,w;
  long p1,p2;
  long ans_p1=-1, ans_p2=-1;
  long min_diff = 0xFFFFFFFFFFFFF;
  int left, mid, right;
  left = 0;
  right = width;
  mid = (left + width) / 2;
  while (left != mid) {
    p1 = p2 = 0;
    for (h=0; h < height; h++) {
      for (w=0; w < width; w++) {
	int count = blackBitsNum(pixel[h * width + w]);
	if (w < mid) {
	  p1 += count;
	} else {
	  p2 += count;
	}
      }
      long diff = labs(p1 - p2);
      if (diff < min_diff) {
	// 線を保存
	line = mid;
	min_diff = diff;
	ans_p1 = p1;
	ans_p2 = p2;
      }
    }
    // 面積が大きい方に線を寄せる
    if (p1 > p2) {
      right = mid;
    } else {
      left = mid;
    }
    mid = (left + right) / 2;
  }
  printf("line = %d\n", mid);
  printf("ans_p1 = %ld\n", ans_p1);
  printf("ans_p2 = %ld\n", ans_p2);
  return pixel;
}

unsigned char* divide2(unsigned char* pixel, int width, int height, int divide)
{
    if (divide == 1) {
        printf("%d", height);
        return pixel;
    }
  width /= 8;
  int w,h,count;
  // すべての黒面積を計算
  int blackPixel = 0;
  for (h=0; h < height; h++) {
    for (w=0; w < width; w++) {
      blackPixel += blackBitsNum(pixel[h * width + w]);
    }
  }

  int i = 0;
    int divideCount = 0;
  blackPixel /= divide;

  for (h=0; h < height; h++) {
    for (w=0; w < width; w++) {  
      count += blackBitsNum(pixel[h * width + w]); 
    }
    if ( count > blackPixel ) {
        divideCount++;
        if (divideCount == divide-1) {
            printf("%d,%d", h, height);
        } else {
      printf("%d,", h);
        }

      // line をシマシマにする
      int k, s=0;
      for (k=0; k < width; k++) {
        pixel[h * width + k] = (s = ~s) ? 0x00 : 0xFF;
      }
      count = 0;
    }
  }
  return pixel;
}

void write_image (file_header *fh, information_header *ih, unsigned char *pixel, int size)
{
  FILE *fp;
  fp = fopen("./image.bmp", "wb");
  if ( fp == NULL ) {
    printf("cant open image,bmp\n");
  }

  fwrite(fh, sizeof(file_header), 1, fp);
  fwrite(ih, sizeof(information_header), 1, fp);
  fwrite(pixel, sizeof(char), size, fp);
  fclose(fp);
}

int main(int argc, char *argv[])
{
  FILE *fp;
  fp = fopen(argv[1], "rb");
  if ( fp == NULL ) {
    printf("can't read %s\n", argv[1]);
    return -1;
  }

  int imgSize = 0;
  
  file_header *f_header;
  f_header = (file_header *)malloc(sizeof(file_header));

  information_header *i_header;
  i_header = (information_header *)malloc(sizeof(information_header));

  imgSize += fread(f_header, sizeof(file_header), 1, fp);
  imgSize += fread(i_header, sizeof(information_header), 1, fp);

  int width = i_header->biWidth;
  int height = i_header->biHeight;
  long pixelSize = width * height;
  imgSize += pixelSize;
  
  unsigned char *pixel;
  // 確保しすぎじゃない?
  pixel = (unsigned char *)malloc( pixelSize / sizeof(unsigned char) );
  fread(pixel, sizeof(char), pixelSize, fp);

  pixel = divide2(pixel, width, height, atoi(argv[2]));
  
  write_image(f_header, i_header, pixel, imgSize);
}
